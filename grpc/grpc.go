package grpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	co "github.com/subiz/header/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CredKey is key which credential is putted in medatada.MD
const (
	CredKey  = "credential"
	CtxKey   = "pcontext"
	ErrKey   = "error"
	PanicKey = "panic"
)

func NewCacheInterceptor() func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var cachedMethods sync.Map

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e8,     // number of keys to track frequency of (100M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := cachedMethods.Load(method); ok {
			key := proto.MarshalTextString(req.(proto.Message))
			if val, ok := cache.Get(key); ok {
				return proto.Unmarshal(val.([]byte), reply.(proto.Message))
			}
		}

		var header metadata.MD // variable to store header and trailer
		opts = append([]grpc.CallOption{grpc.Header(&header)}, opts...)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return err
		}

		if len(header["max-age"]) > 0 {
			maxage, err := strconv.Atoi(header["max-age"][0])
			if err == nil {
				key := proto.MarshalTextString(req.(proto.Message))
				val, _ := proto.Marshal(reply.(proto.Message))
				if maxage > 0 {
					cachedMethods.Store(method, true)
					cache.SetWithTTL(key, val, 1, time.Duration(maxage)*time.Second)
				} else {
					cache.Del(key)
				}
			}
		}

		return nil
	}
}

// SetMaxAge is used by the grpc server to tell clients the response isn't going
// to change for the next some seconds, and its safe to reuse this response
func SetMaxAge(ctx context.Context, sec int) {
	header := metadata.Pairs("max-age", fmt.Sprintf("%d", sec))
	grpc.SendHeader(ctx, header)
}

func ToGrpcCtx(pctx *co.Context) context.Context {
	data, err := proto.Marshal(pctx)
	if err != nil {
		panic(fmt.Sprintf("unable to marshal cred, %v", pctx))
	}
	cred64 := base64.StdEncoding.EncodeToString(data)
	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs(CtxKey, cred64))
}

func FromGrpcCtx(ctx context.Context) *co.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md, ok = metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil
		}
	}
	cred64 := strings.Join(md[CtxKey], "")
	if cred64 == "" {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(cred64)
	if err != nil {
		panic(fmt.Sprintf("%v, %s: %s", err, "wrong base64 ", cred64))
	}

	pctx := &co.Context{}
	if err = proto.Unmarshal(data, pctx); err != nil {
		panic(fmt.Sprintf("%v, %s: %s", err, "unable to unmarshal cred ", cred64))
	}
	return pctx
}

func unaryinterceptorhandler(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (ret interface{}, err error) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if ok {
					err = errors.Wrap(e, 500, errors.E_unknown)
				}

				err = errors.New(500, errors.E_unknown, fmt.Sprintf("%v", e))
			}
		}()
		ret, err = handler(ctx, req)
	}()
	if err != nil {
		e, ok := err.(*errors.Error)
		if !ok {
			e, _ = errors.Wrap(err, 500, errors.E_unknown).(*errors.Error)
		}
		md := metadata.Pairs(PanicKey, e.Error())
		grpc.SendHeader(ctx, md)
	}
	return ret, err
}

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func NewRecoveryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(unaryinterceptorhandler)
}

func GetPanic(md metadata.MD) *errors.Error {
	errs := strings.Join(md[PanicKey], "")
	if errs == "" {
		return nil
	}
	return errors.FromString(errs)
}

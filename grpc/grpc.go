package grpc

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	co "github.com/subiz/header/common"
	"golang.org/x/net/context"
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

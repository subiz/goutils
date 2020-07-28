package grpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	co "github.com/subiz/header/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	protoV2 "google.golang.org/protobuf/proto"
)

// CredKey is key which credential is putted in medatada.MD
const (
	CredKey  = "credential"
	CtxKey   = "pcontext"
	ErrKey   = "error"
	PanicKey = "panic"
)

func NewShardInterceptor() func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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

// forward proxy a GRPC calls to another host, header and trailer are preserved
// parameters:
//   host: host address which will be redirected to
//   method: the full RPC method string, i.e., /package.service/method.
//   returnedType: type of returned value
//   in: value of input (in request) parameter
// this method returns output just like a normal GRPC call
func (me *ForwardServer) forward(host, method string, returnedType reflect.Type, ctx context.Context, in interface{}, extraHeader metadata.MD) (interface{}, error) {
	// use cache host connection or create a new one
	me.Lock()
	cc, ok := me.conn[host]
	if !ok {
		var err error
		cc, err = dialGrpc(host)
		if err != nil {
			me.Unlock()
			return nil, err
		}
		me.conn[host] = cc
		me.Unlock()
	}

	md, _ := metadata.FromIncomingContext(ctx)
	extraHeader = metadata.Join(extraHeader, md)
	outctx := metadata.NewOutgoingContext(context.Background(), extraHeader)

	out := reflect.New(returnedType).Interface()
	var header, trailer metadata.MD
	err := cc.Invoke(outctx, method, in, out, grpc.Header(&header), grpc.Trailer(&trailer))
	grpc.SendHeader(ctx, header)
	grpc.SetTrailer(ctx, trailer)

	return out, err
}

// dialGrpc makes a GRPC connection to service
func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(5*time.Second))
	return grpc.Dial(service, opts...)
}

type ForwardServer struct {
	*sync.Mutex
	conn map[string]*grpc.ClientConn
}

// NewShardIntercept makes a new GRPC shard server interceptor
func NewShardIntercept(serviceAddrs []string, id int) grpc.ServerOption {
	numShard := len(serviceAddrs)
	analysed := false
	var returnedTypeM map[string]reflect.Type

	lock := &sync.Mutex{}
	s := &ForwardServer{
		conn:  make(map[string]*grpc.ClientConn),
		Mutex: &sync.Mutex{},
	}

	f := func(ctx context.Context, in interface{}, sinfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		lock.Lock()
		if !analysed {
			returnedTypeM = analysisReturnType(sinfo.Server)
			analysed = true
		}
		lock.Unlock()

		md, _ := metadata.FromIncomingContext(ctx)
		pkey := strings.Join(md["shard_key"], "")
		if pkey == "" {
			pkey = GetAccountId(in)
			if pkey == "" {
				// no sharding parameter, perform the request anyway
				return handler(ctx, in)
			}
		}

		lock.Lock()
		ghash.Write([]byte(pkey))
		parindex := ghash.Sum32() % uint32(numShard) // 1024
		ghash.Reset()
		lock.Unlock()

		if int(parindex) == id { // correct partition
			return handler(ctx, in)
		}

		extraHeader := metadata.New(nil)
		// the request have been proxied and then proxied, ...
		// we give up to prevent looping
		redirectOfRedirect := len(strings.Join(md["shard_redirected_2"], "")) > 0
		if redirectOfRedirect {
			return nil, status.Errorf(codes.Internal, "Sharding inconsistent")
		}

		// the request just have been proxied and still
		// doesn't arrived to the correct host
		// this happend when total_shards is not consistent between servers. We will wait for
		// 5 secs and then proxy one more time. Hoping that the consistent will be resolved
		justRedirect := len(strings.Join(md["shard_redirected"], "")) > 0
		if justRedirect {
			extraHeader.Set("shard_redirected_2", "true")
			time.Sleep(5 * time.Second)
		} else {
			extraHeader.Set("shard_redirected", "true")
		}

		// the correct host
		methodSplit := strings.Split(sinfo.FullMethod, "/")
		shortmethod := methodSplit[len(methodSplit)-1]
		header := metadata.Pairs("total_shards", strconv.Itoa(numShard))
		grpc.SendHeader(ctx, header)
		return s.forward(serviceAddrs[parindex], sinfo.FullMethod, returnedTypeM[shortmethod], ctx, in, extraHeader)
	}

	return grpc.UnaryInterceptor(f)
}

// analysisReturnType returns all return types for every GRPC method in server handler
// the returned map takes full method name (i.e., /package.service/method) as key, and the return type as value
// For example, with handler
//   (s *server) func Goodbye() string {}
//   (s *server) func Ping(_ context.Context, _ *pb.Ping) (*pb.Pong, error) {}
//   (s *server) func Hello(_ context.Context, _ *pb.Empty) (*pb.String, error) {}
// this function detected 2 GRPC methods is Ping and Hello, it would return
// {"Ping": *pb.Pong, "Hello": *pb.Empty}
func analysisReturnType(server interface{}) map[string]reflect.Type {
	m := make(map[string]reflect.Type)
	t := reflect.TypeOf(server)
	for i := 0; i < t.NumMethod(); i++ {
		methodType := t.Method(i).Type
		if methodType.NumOut() != 2 || methodType.NumIn() < 2 {
			continue
		}

		if methodType.In(2).Kind() != reflect.Ptr || methodType.In(1).Name() != "Context" {
			continue
		}

		if methodType.Out(0).Kind() != reflect.Ptr || methodType.Out(1).Name() != "error" {
			continue
		}

		m[t.Method(i).Name] = methodType.Out(0).Elem()
	}
	return m
}

func GetAccountId(message interface{}) string {
	msgrefl := message.(protoV2.Message).ProtoReflect()
	accIdDesc := msgrefl.Descriptor().Fields().ByName("account_id")
	if accIdDesc == nil {
		return ""
	}
	return msgrefl.Get(accIdDesc).String()
}

// global hashing util, used to hash key to partition number
var ghash = fnv.New32a()

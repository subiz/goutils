package grpc

import (
	"git.subiz.net/errors"
	co "git.subiz.net/header/common"
	"git.subiz.net/header/lang"
	b64 "encoding/base64"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
	"fmt"
)

// CredKey is key which credential is putted in medatada.MD
const (
	CredKey      = "credential"
	CtxKey       = "pcontext"
	ErrKey       = "error"
	RedirectKey  = "isredirect"
	RequestIDKey = "requestid"
	PanicKey     = "panic"
)

func AddRequestIDCtx(ctx context.Context, requestid string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md, ok = metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
	}

	md = metadata.Join(metadata.Pairs(RequestIDKey, requestid), md)
	return metadata.NewOutgoingContext(ctx, md)
}

func ToGrpcCtx(pctx *co.Context) context.Context {
	data, err := proto.Marshal(pctx)
	if err != nil {
		panic(fmt.Sprintf("unable to marshal cred, %v", pctx))
	}
	cred64 := b64.StdEncoding.EncodeToString(data)
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
	data, err := b64.StdEncoding.DecodeString(cred64)
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
				ok := false
				if err, ok = r.(error); !ok {
					err = errors.New(500, lang.T_internal_error, r)
				}
			}
		}()
		ret, err = handler(ctx, req)
	}()
	if err != nil {
		e, ok := err.(*errors.Error)
		if !ok {
			e = errors.Wrap(err, 500, lang.T_internal_error)
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
	return errors.FromError(errs)
}

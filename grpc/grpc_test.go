package grpc

import (
	"context"
	"github.com/subiz/header"
	pb "github.com/subiz/header/api"
	"google.golang.org/grpc"
	"net"
	"testing"
	"google.golang.org/grpc/metadata"
	"time"
)

type TestCacheApiServer struct {
	ncall int
}

func (me *TestCacheApiServer) Call(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	me.ncall++
	if me.ncall > 3 {
		// insufficient caching
		return &pb.Response{Code: 500}, nil
	}
	SetMaxAge(ctx, 2)
	return &pb.Response{Code: 212}, nil
}

func TestCache(t *testing.T) {
	lis, err := net.Listen("tcp", ":21234")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	header.RegisterApiServerServer(grpcServer, &TestCacheApiServer{})
	go grpcServer.Serve(lis)

	conn, err := dialGrpc(":21234")
	if err != nil {
		panic(err)
	}
	client := header.NewApiServerClient(conn)

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		var header metadata.MD // variable to store header and trailer
		r, _ := client.Call(context.Background(), &pb.Request{Method: "thanh"}, grpc.Header(&header))

		if r.Code != 212 {
			t.Fatal("SHOULD BE 212")
		}
	}
}

func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(120*time.Second), grpc.WithUnaryInterceptor(NewCacheInterceptor()))
	return grpc.Dial(service, opts...)
}

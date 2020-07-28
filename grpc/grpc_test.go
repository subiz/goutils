package grpc

import (
	// "fmt"
	"strings"
	//	proto "github.com/golang/protobuf/proto"
	"context"
	"github.com/subiz/header"
	pb "github.com/subiz/header/api"
	cpb "github.com/subiz/header/common"
	upb "github.com/subiz/header/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"strconv"
	"testing"
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

func (me *TestCacheApiServer) Serve() {
	lis, err := net.Listen("tcp", ":21234")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	header.RegisterApiServerServer(grpcServer, me)
	grpcServer.Serve(lis)
}

func TestCache(t *testing.T) {
	server := &TestCacheApiServer{}
	go server.Serve()

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

type TestShardApiServer struct {
	id int
}

func (me *TestShardApiServer) ListTopVisitors(ctx context.Context, req *cpb.Id) (*upb.Visitors, error) {
	return &upb.Visitors{}, nil
}

func (me TestShardApiServer) Serve(id int) {
	lis, err := net.Listen("tcp", ":2124"+strconv.Itoa(id))
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer(NewShardIntercept([]string{":21240", ":21241"}, id))
	header.RegisterVisitorMgrServer(grpcServer, &me)
	grpcServer.Serve(lis)
}

func TestShard(t *testing.T) {
	server0 := &TestShardApiServer{}
	go server0.Serve(0)

	server1 := &TestShardApiServer{}
	go server1.Serve(1)

	conn, err := dialGrpc(":21240")
	if err != nil {
		panic(err)
	}
	client := header.NewVisitorMgrClient(conn)

	// correct server
	var header metadata.MD // variable to store header and trailer
	client.ListTopVisitors(context.Background(), &cpb.Id{AccountId: "thanh"}, grpc.Header(&header))
	if strings.Join(header.Get("shard_num"), "") != "" {
		t.Fatal("SHOULD NOT RETURN ANY SHARD_NUM")
	}

	// must redirect
	var header2 metadata.MD // variable to store header and trailer
	client.ListTopVisitors(context.Background(), &cpb.Id{AccountId: "thanh1"}, grpc.Header(&header2))
	if strings.Join(header2.Get("total_shards"), "") != "2" {
		t.Fatal("SHOULD RETURN SHARD NUM", strings.Join(header2.Get("total_shards"), ""))
	}
}

func TestProto(t *testing.T) {
	msg := &pb.WhitelistUrl{AccountId: "thanh"}
	if "thanh" != GetAccountId(msg) {
		t.Errorf("SHOULD BE THANH, GOT %s", GetAccountId(msg))
	}
}

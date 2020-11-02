package main

import (
	"context"
	"discovery/discovery_go_grpc/grpclb"
	greetProto "discovery/discovery_go_grpc/proto"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	s := &grpclb.ServiceInfo{
		RegisterDir: "service",
		Name:        "greet",
		Version:     "v1",
		Endpoint:    "127.0.0.1:9001",
		TTL:         15,
	}
	lis, err := net.Listen("tcp", s.Endpoint)
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	registrar, err := grpclb.NewRegistrar()
	if err != nil {
		log.Fatalf("grpclb new registrar err: %s", err.Error())
	}
	go registrar.Register(s, clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	greetProto.RegisterGreetServiceServer(grpcServer, Handler{})
	grpcServer.Serve(lis)
}

type Handler struct{}

func (h Handler) Hi(ctx context.Context, req *greetProto.HiRequest) (*greetProto.HiResponse, error) {
	res := &greetProto.HiResponse{
		Message: fmt.Sprintf("Hi, %s", req.Name),
	}
	return res, nil
}

package main

import (
	"context"
	greetProto "discovery/discovery-go-grpc/proto"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:9001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	greetProto.RegisterGreetServiceServer(grpcServer, Handler{})
	grpcServer.Serve(lis)
}

type Handler struct {}

func (h Handler) Hi(ctx context.Context, req *greetProto.HiRequest) (*greetProto.HiResponse, error) {
	res := &greetProto.HiResponse{
		Message: fmt.Sprintf("Hi, %s", req.Name),
	}
	return res, nil
}
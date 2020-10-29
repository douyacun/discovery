package main

import (
	"context"
	greetProto "discovery/discovery-go-grpc/proto"
	"fmt"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:9001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("err: %v", err.Error())
		return
	}
	defer conn.Close()
	client := greetProto.NewGreetServiceClient(conn)
	resp, err := client.Hi(context.TODO(), &greetProto.HiRequest{Name: "刘宁"})
	if err != nil {
		log.Fatalf("call hi err: %s", err.Error())
		return
	}
	fmt.Println(resp.Message)
}

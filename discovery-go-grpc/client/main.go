package main

import (
	"context"
	"discovery/discovery-go-grpc/grpclb"
	greetProto "discovery/discovery-go-grpc/proto"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	b := grpc.RoundRobin(grpclb.NewEtcdResolver("service", "greet", cli))
	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithBalancer(b), grpc.WithBlock())
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
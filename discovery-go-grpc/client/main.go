package main

import (
	"context"
	greetProto "discovery/discovery-go-grpc/proto"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc"
	"log"
	"time"
)
// https://colobu.com/2017/03/25/grpc-naming-and-load-balance/#2%EF%BC%89%E6%9C%8D%E5%8A%A1%E5%8F%91%E7%8E%B0%E5%AE%9E%E7%8E%B0%EF%BC%9Awatcher-go
func main() {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("connect etcd %s failed", "127.0.0.1:2379")
	}
	r := &etcdnaming.GRPCResolver{Client: etcdClient}
	b := grpc.RoundRobin(r)
	conn, err := grpc.Dial("localhost:9001", grpc.WithInsecure(), grpc.WithBalancer(b))
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

type resolver struct {
	serviceName string
}

func watch() {

}
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		log.Println(err)
	}
	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		log.Print(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := cli.KeepAlive(ctx, resp.ID)
	if err != nil {
		log.Println(err)
	}
	go func() {
		for v := range ch {
			fmt.Printf("%d ttl %d\n", v.ID, v.TTL)
		}
	}()
	time.Sleep(60 * time.Second)
}

package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	if err := watch(); err != nil {
		fmt.Printf("err: %s", err.Error())
	}
}

func watch() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}

	service := "/discovery_service"
	ch := cli.Watch(context.TODO(), service, clientv3.WithPrefix())
	for wresp := range ch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				fmt.Printf("put %s %s\n", ev.Kv.Key, ev.Kv.Value)
			case mvccpb.DELETE:
				fmt.Printf("delete %s %s\n", ev.Kv.Key, ev.Kv.Value)
			}
		}
	}
	return nil
}

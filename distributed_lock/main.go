package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"log"
	"sync"
	"time"
)

var i = 0

func Do(key string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		return err
	}
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return err
	}
	m := concurrency.NewMutex(session, fmt.Sprintf("/%s", key))
	if err := m.Lock(context.TODO()); err != nil {
		return err
	}
	defer func() {
		if err := m.Unlock(context.TODO()); err != nil {
			log.Printf("err: %s", err.Error())
		}
	}()
	log.Printf("get lock: %d", i)
	i++
	time.Sleep(time.Second * 3)
	return nil
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			if err := Do("lock"); err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Wait()
}

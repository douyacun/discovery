package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"net"
	"time"
)

func main() {
	if err := register(); err != nil {
		fmt.Printf("err: %s", err.Error())
	}
}

func getClientIp() (string, error) {
	addrNets, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, v := range addrNets {
		if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	panic("unable to determine local ip")
}

func register() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}

	ip, err := getClientIp()
	if err != nil {
		return err
	}
	service := "/discovery_service"
	_, err = cli.Put(context.TODO(), fmt.Sprintf("%s/%s", service, ip), ip, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	for{
		ka, err := cli.KeepAliveOnce(context.TODO(), resp.ID)
		if err != nil {
			log.Printf("err: %s", err.Error())
		}
		fmt.Println("续租成功 ttl: ", ka.TTL)
		time.Sleep(3 * time.Second)
	}
}

package main

import (
	"context"
	greetProto "discovery/discovery-go-grpc/proto"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

func main() {
	s := &Service{
		Endpoints:     "localhost:9001",
		EtcdEndpoints: []string{"127.0.0.1:2379"},
		TTL:           15,
		Prefix:        "discover-service",
		ServiceName:   "greet",
	}
	s.Register(context.TODO())
	lis, err := net.Listen("tcp", s.Endpoints)
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	greetProto.RegisterGreetServiceServer(grpcServer, Handler{})
	grpcServer.Serve(lis)
}

type Service struct {
	Endpoints     string
	EtcdEndpoints []string
	TTL           int64
	Prefix        string
	ServiceName   string
}

func (s *Service) Register(ctx context.Context) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   s.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	serviceKey := fmt.Sprintf("/%s/%s/%s", s.Prefix, s.ServiceName, s.ServiceName)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			resp, _ := client.Grant(ctx, int64(15))
			_, err := client.Get(ctx, serviceKey)
			if err != nil {
				if err == rpctypes.ErrKeyNotFound {
					if _, err := client.Put(ctx, serviceKey, s.Endpoints, clientv3.WithLease(resp.ID)); err != nil {
						log.Fatalf("grpc set service '%s' with ttl to etcd failed: %s", s.ServiceName, err.Error())
					}
				} else {
					log.Fatalf("grpc service '%s' connect to etcd v3 failed: %s", s.ServiceName, err.Error())
				}
			} else {
				if _, err := client.Put(ctx, serviceKey, s.Endpoints, clientv3.WithLease(resp.ID)); err != nil {
					log.Fatalf("grpc refresh service '%s' with ttl to etcd v3 faield: %s", s.ServiceName, err.Error())
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

type Handler struct{}

func (h Handler) Hi(ctx context.Context, req *greetProto.HiRequest) (*greetProto.HiResponse, error) {
	res := &greetProto.HiResponse{
		Message: fmt.Sprintf("Hi, %s", req.Name),
	}
	return res, nil
}

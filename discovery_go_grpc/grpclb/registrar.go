package grpclb

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"path"
	"time"
)

type Registrar struct {
	EtcdClient *clientv3.Client
	Cancel     context.CancelFunc
}

func NewRegistrar() (*Registrar, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Registrar{
		EtcdClient: cli,
	}, nil
}

func (r *Registrar) Register(service *ServiceInfo, config clientv3.Config) error {
	val, err := json.Marshal(service)
	if err != nil {
		return err
	}
	etcdClient, err := clientv3.New(config)
	if err != nil {
		return err
	}
	serviceKey := path.Join("/", service.RegisterDir, service.Name, service.Version, service.Endpoint)
	ctx, cancel := context.WithCancel(context.Background())
	r.Cancel = cancel
	resp, err := etcdClient.Grant(ctx, service.TTL)
	if err != nil {
		grpclog.Errorf("[register]: %s\n", err.Error())
		return err
	}
	if _, err := etcdClient.Put(ctx, serviceKey, string(val), clientv3.WithLease(resp.ID)); err != nil {
		grpclog.Errorf("[register]: %s\n", err.Error())
		return err
	}
	ticker := time.NewTicker(time.Duration(int(service.TTL/2)) * time.Second)
	for {
		if _, err := etcdClient.KeepAliveOnce(ctx, resp.ID); err != nil {
			grpclog.Errorf("[register]: %s", err.Error())
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (r *Registrar) Done() {
	r.Cancel()
}

func (r *Registrar) Close() error {
	return r.EtcdClient.Close()
}

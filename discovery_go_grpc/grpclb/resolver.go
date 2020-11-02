package grpclb

import (
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/naming"
	"path"
)

type EtcdResolver struct {
	registerDir string
	serviceName string
	client      *clientv3.Client
}

func NewEtcdResolver(registerDir, serviceName string, client *clientv3.Client) naming.Resolver {
	if registerDir == "" || serviceName == "" {
		panic("[resolver]: 服务注册前缀或服务名称不能为空")
	}
	return &EtcdResolver{
		registerDir: registerDir,
		serviceName: serviceName,
		client:      client,
	}
}

func (r *EtcdResolver) Resolve(target string) (naming.Watcher, error) {
	key := path.Join("/", r.registerDir, r.serviceName)
	return NewEtcdWatcher(r.client, key), nil
}

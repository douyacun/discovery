package grpclb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
)

var ErrWatcherClosed = fmt.Errorf("naming: watch closed")

type EtcdWatcher struct {
	key    string
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
	wc     clientv3.WatchChan
}

func NewEtcdWatcher(cli *clientv3.Client, target string) naming.Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &EtcdWatcher{
		key:    target,
		client: cli,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *EtcdWatcher) Next() ([]*naming.Update, error) {
	if w.wc == nil {
		return w.FirstNext()
	}
	wr, ok := <-w.wc
	if !ok {
		return nil, ErrWatcherClosed
	}
	updates := make([]*naming.Update, 0)
	for _, event := range wr.Events {
		switch event.Type {
		case clientv3.EventTypeDelete:
			var serviceInfo ServiceInfo
			if err := json.Unmarshal(event.Kv.Value, &serviceInfo); err != nil {
				return nil, err
			}
			grpclog.Errorf("[watcher]: delete endpoint %s", serviceInfo.Endpoint)
			updates = append(updates, &naming.Update{
				Op:   naming.Delete,
				Addr: serviceInfo.Endpoint,
			})
		case clientv3.EventTypePut:
			var serviceInfo ServiceInfo
			if err := json.Unmarshal(event.Kv.Value, &serviceInfo); err != nil {
				return nil, err
			}
			grpclog.Errorf("[watcher]: add endpoint %s", serviceInfo.Endpoint)
			updates = append(updates, &naming.Update{
				Op:   naming.Add,
				Addr: serviceInfo.Endpoint,
			})
		}
	}
	return updates, nil
}

func (w *EtcdWatcher) FirstNext() ([]*naming.Update, error) {
	resp, err := w.client.Get(w.ctx, w.key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	updates := make([]*naming.Update, 0)
	for _, v := range resp.Kvs {
		var serviceInfo ServiceInfo
		if err := json.Unmarshal(v.Value, &serviceInfo); err != nil {
			grpclog.Errorf("[watcher]: %s", err.Error())
			continue
		}
		grpclog.Infof("[watcher]: %v\n", serviceInfo)
		updates = append(updates, &naming.Update{
			Op:   naming.Add,
			Addr: serviceInfo.Endpoint,
		})
	}
	w.wc = w.client.Watch(w.ctx, w.key, clientv3.WithPrefix())
	return updates, nil
}

func (w *EtcdWatcher) Close() {
	w.cancel()
	_ = w.client.Close()
}

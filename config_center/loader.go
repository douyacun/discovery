package config_center

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"
)

type Loader struct {
	Name         string
	Filepath     string
	EtcdEndpoint []string
	Config       map[string]interface{}
}

func (l *Loader) LoadFileConfig() error {
	data, err := ioutil.ReadFile(l.Filepath)
	if err != nil {
		return errors.Errorf("[loader] path: %s err: %s", l.Filepath, err.Error())
	}
	configMap := make(map[string]interface{})
	ext := filepath.Ext(l.Filepath)
	switch ext {
	case ".yaml":
		if err := yaml.Unmarshal(data, &configMap); err != nil {
			return errors.Errorf("[loader] path: %s err: %s", l.Filepath, err.Error())
		}
	case ".json":
		if err := json.Unmarshal(data, &configMap); err != nil {
			return errors.Errorf("[loader] path: %s err: %s", l.Filepath, err.Error())
		}
	default:
		return errors.Errorf("[loader] nonsupport extension %s", ext)
	}

	for k, v := range configMap {
		if conf, ok := l.Config[k]; ok {
			if err := mapstructure.Decode(v, conf); err != nil {
				return errors.Errorf("[loader] map to struct err: %s", err.Error())
			}
		}
	}
	// todo load config watcher
	return nil
}

func (l *Loader) LoadEtcdConfig() error {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return errors.Errorf("[loader] client connect err: %s", err.Error())
	}
	prefixKey := path.Join("/", "config", l.Name)
	ctx, _ := context.WithTimeout(context.Background(), 3 * time.Second)
	response, err := etcdClient.Get(ctx, prefixKey, clientv3.WithPrefix())
	if err != nil {
		return errors.Errorf("[loader] get %s err: %s", prefixKey, err.Error())
	}
	if response.Count == 0 {
		return errors.Errorf("[loader] etcd get %s empty", prefixKey)
	}
}

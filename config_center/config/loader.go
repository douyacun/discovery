package config

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Load struct {
	Name         string
	Filepath     string
	EtcdEndpoint []string
	Config       map[string]interface{}
}

const FormatKey = "__FORMAT__"

func (l *Load) LoadFileConfig() error {
	configMap, err := l.load(l.Filepath)
	if err != nil {
		return err
	}
	for k, v := range configMap {
		if conf, ok := l.Config[k]; ok {
			if err := mapstructure.Decode(v, conf); err != nil {
				return errors.Errorf("[config load] map to struct err: %s", err.Error())
			}
		}
	}
	// todo load config watcher
	return nil
}

func (l *Load) LoadEtcdConfig() error {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return errors.Errorf("[config load] client connect err: %s", err.Error())
	}
	prefixKey := path.Join("/", "config", l.Name)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	resp, err := etcdClient.Get(ctx, prefixKey, clientv3.WithPrefix())
	if err != nil {
		return errors.Errorf("[config load] get %s err: %s", prefixKey, err.Error())
	}
	if resp.Count == 0 {
		return errors.Errorf("[config load] etcd get %s empty", prefixKey)
	}
	var ext string
	for _, item := range resp.Kvs {
		if strings.Contains(string(item.Key), FormatKey) {
			ext = string(item.Value)
		}
	}
	for _, item := range resp.Kvs {
		keys := strings.Split(string(item.Key), "/")
		key := keys[len(keys)-1:][0]
		if key == FormatKey {
			continue
		}
		if config, ok := l.Config[key]; ok {
			switch ext {
			case ".json":
				if err := json.Unmarshal(item.Value, config); err != nil {
					return errors.Errorf("[config load] json unmarshal err: %s", err.Error())
				}
			case ".yaml":
				if err := yaml.Unmarshal(item.Value, config); err != nil {
					return errors.Errorf("[config load] yaml unmarshal err: %s", err.Error())
				}
			}
		}
	}
	go func() {
		watchCh := etcdClient.Watch(context.TODO(), prefixKey, clientv3.WithPrefix())
		for resp := range watchCh {
			for _, e := range resp.Events {
				keys := strings.Split(string(e.Kv.Key), "/")
				key := keys[len(keys)-1:][0]
				if conf, ok := l.Config[key]; ok{
					switch ext {
					case ".json":
						if err := json.Unmarshal(e.Kv.Value, conf); err != nil {
							return
						}
					case ".yaml":
						if err := yaml.Unmarshal(e.Kv.Value, conf); err != nil {
							return
						}
					}
				}
			}
		}
	}()
	return nil
}

func (l *Load) PushEtcdConfig() error {
	configMap, err := l.load(l.Filepath)
	if err != nil {
		return err
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	for k, v := range configMap {
		data := ""
		if k != FormatKey {
			switch configMap[FormatKey] {
			case ".json":
				if b, err := json.MarshalIndent(v, "", " "); err != nil {
					return errors.Errorf("[config load] json marshal err: %s", err.Error())
				} else {
					data = string(b)
				}
			case ".yaml":
				if b, err := yaml.Marshal(v); err != nil {
					return errors.Errorf("[config load] yaml marshal err: %s", err.Error())
				} else {
					data = string(b)
				}
			}
		} else {
			data = v.(string)
		}
		key := path.Join("/", "config", l.Name, k)
		if _, err := client.Put(context.TODO(), key, data); err != nil {
			return errors.Errorf("[config load] etcd put %s %s err: %s", key, data, err.Error())
		}
	}
	return nil
}

func (l *Load) load(path string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("[config load] read file %s err: %s", path, err.Error())
	}
	configMap := make(map[string]interface{})
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml":
		if err := yaml.Unmarshal(data, &configMap); err != nil {
			return nil, errors.Errorf("[config load] yaml unmarshal error: %s", err.Error())
		}
	case ".json":
		if err := json.Unmarshal(data, &configMap); err != nil {
			return nil, errors.Errorf("[config load] json unmarshal error: %s", err.Error())
		}
	default:
		return nil, errors.Errorf("[config load] nonsupport config file extension")
	}
	configMap[FormatKey] = ext
	return configMap, nil
}

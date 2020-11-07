package main

import (
	"etcd/config_center/config"
	"fmt"
	"log"
	"time"
)

type EnvConfig struct {
	Env string `json:"Env"`
}

func main() {
	LoadConfig()
}

func LoadConfig() {
	configMap := make(map[string]interface{})
	configMap["EnvConfig"] = &EnvConfig{}
	l := &config.Load{
		Name:         "config_center",
		Filepath:     "./config.json",
		EtcdEndpoint: []string{"127.0.0.1:2379"},
		Config:       configMap,
	}
	if err := l.LoadEtcdConfig(); err != nil {
		log.Println(err)
		return
	}
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			envConfig := configMap["EnvConfig"].(*EnvConfig)
			fmt.Printf("env: %s\n", envConfig.Env)
		}
	}
}

func PushConfig() {
	configMap := make(map[string]interface{})
	configMap["EnvConfig"] = &EnvConfig{}
	l := &config.Load{
		Name:         "config_center",
		Filepath:     "./config.json",
		EtcdEndpoint: []string{"127.0.0.1:2379"},
		Config:       configMap,
	}
	if err := l.PushEtcdConfig(); err != nil {
		log.Println(err)
		return
	}
}

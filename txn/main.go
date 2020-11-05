package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"strconv"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		log.Println(err)
		return
	}

	// 经典转账
	from := "a"
	to := "b"
	amount := int64(100)
	// 初始化账户
	if _, err := client.Txn(context.TODO()).If(clientv3.Compare(clientv3.CreateRevision(from), "=", 0)).Then(clientv3.OpPut(from, "10000")).Commit(); err != nil {
		log.Println(err)
		return
	}
	if _, err := client.Txn(context.TODO()).If(clientv3.Compare(clientv3.CreateRevision(to), "=", 0)).Then(clientv3.OpPut(to, "2000")).Commit(); err != nil {
		log.Println(err)
		return
	}
	// 获取账户余额
	getResp, err := client.Txn(context.TODO()).Then(clientv3.OpGet(from), clientv3.OpGet(to)).Commit()
	if err != nil {
		log.Println(err)
		return
	}
	fromKV := getResp.Responses[0].GetResponseRange().Kvs[0]
	toKV := getResp.Responses[1].GetResponseRange().Kvs[0]
	fromBalance, err := strconv.ParseInt(string(fromKV.Value), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}
	toBalance, err := strconv.ParseInt(string(toKV.Value), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}
	// 余额充足
	if fromBalance < amount {
		fmt.Printf("余额不足, 当前余额: %d\n", fromBalance)
		return
	}
	// 转账
	putResp, err := client.Txn(context.TODO()).If(
		clientv3.Compare(clientv3.ModRevision(from), "=", fromKV.ModRevision),
		clientv3.Compare(clientv3.ModRevision(to), "=", toKV.ModRevision)).
		Then(clientv3.OpPut(from, strconv.FormatInt(fromBalance-amount, 10)),
			clientv3.OpPut(to, strconv.FormatInt(toBalance+amount, 10))).Commit()
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("转账状态: %v\n", putResp.Succeeded)
	return
}

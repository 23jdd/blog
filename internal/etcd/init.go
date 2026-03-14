package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdClient struct {
	*clientv3.Client
}

var client *clientv3.Client

func NewEtcdClient() *EtcdClient {
	return &EtcdClient{
		Client: client,
	}
}
func init() {
	var err error
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(fmt.Sprintf("Error connecting to etcd: %v", err))
	}
}
func (c *EtcdClient) Close() error {
	return c.Client.Close()
}

// 分两步实现加锁
// 1. 申请一个租约，10秒过期
// 2. 尝试在租约下创建一个键值对，如果键值对不存在，则加锁成功
// 加锁
func (e *EtcdClient) TryLock(key string) bool {
	// 申请一个租约，10秒过期
	lease, err := e.Client.Grant(context.Background(), 100)
	if err != nil {
		panic(fmt.Sprintf("Error granting lease: %v", err))
	}
	txn := e.Client.Txn(context.Background())
	txn.If(clientv3.Compare(clientv3.Version(key), "=", 0))
	txn.Then(clientv3.OpPut(key, "", clientv3.WithLease(lease.ID)))
	txn.Else()
	resp, err := txn.Commit()
	if err != nil {
		panic(fmt.Sprintf("Error committing transaction: %v", err))
	}
	if resp.Succeeded {
		return true
	}
	fmt.Println("Lock acquisition failed")
	return false
}

// 解锁
func (c *EtcdClient) Unlock(key string) {
	c.Client.Delete(context.Background(), key)
}

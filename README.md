etcd 的特点

- 简单： 基于HTTP+JSON的API让你可以用CURL命令就可以轻松使用。
- 安全： 可以选择SSL客户认证机制。
- 快速： 每个实例每秒支持一千次写操作。
- 可信： 使用Ralf算法充分实现了分布式。

etcd应用场景

# 服务发现

**服务注册：**

1. 申请lease租约，设置服务生存周期ttl，让key自动过期
2. 服务正常状态下，通过keep-alive定期去租约，避免过期
3. key有统一前缀，比如：`etcdctl --endpoints=127.0.0.1:2379 put /sdmaster-service/192.168.31.110   192.168.31.110`

**服务发现：**

1. 服务`watch /sdmaster-service --prefix`
2. `sdmaster-service` 下发生变化都会发生通知

# grpc 服务发现 负载均衡

正常http负载均衡策略，client -> proxy gateway load balance -> service

grpc提供负载均衡策略，client  load balance -> service 

grpc关键点是负载均衡在客户端

1. 服务端启动时，首先将服务地址注册到服务注册表，定期keepalive健康检测
2. 客户端访问某个服务，watch订阅服务注册表，然后以某种负载均衡策略选择一个目标地址

# 配置中心
- 动态加载配置（watch）
- 历史版本（revision）
- 权限控制
- 命名空间

思路:

1. 加载配置文件 put etcd
2. watch prefix 配置变更可以及时变更配置

# 分布式锁
原理:

1. 租约创建一个key
2. key不存在，创建key，成功获取到锁
3. key存在，无法创建key，获取锁失败

etcd分布式锁的实现在`go.etcd.io/etcd/clientv3/concurrency`包中

1. 初始化sessions

```go
func NewSession(client *v3.Client, opts ...SessionOption) (*Session, error) {
	ops := &sessionOptions{ttl: defaultSessionTTL, ctx: client.Ctx()}
	for _, opt := range opts {
		opt(ops)
	}
	// 初始化租约
	id := ops.leaseID
	if id == v3.NoLease {
		resp, err := client.Grant(ops.ctx, int64(ops.ttl))
		if err != nil {
			return nil, err
		}
		id = v3.LeaseID(resp.ID)
	}
	// 由上层调用控制是否中断
	ctx, cancel := context.WithCancel(ops.ctx)
	keepAlive, err := client.KeepAlive(ctx, id)
	if err != nil || keepAlive == nil {
		cancel()
		return nil, err
	}
	donec := make(chan struct{})
	s := &Session{client: client, opts: ops, id: id, cancel: cancel, donec: donec}
	go func() {
		defer close(donec)
		for range keepAlive {
			// eat messages until keep alive channel closes
		}
	}()
	return s, nil
}
```

2. `func NewMutex(s *Session, pfx string) *Mutex `

初始化`Mutex`

3. `func (m *Mutex) Lock(ctx context.Context) error`

这里类似redis的分布式锁: `set key value NX EX 60`  如果key存在就返回，否则就设置

```go
func (m *Mutex) Lock(ctx context.Context) error {
	s := m.s
	client := m.s.Client()
  // 伪代码
  // if !m.myKey {
  //	client.Put(m.myKey, value)
  // } else {
  // 	client.Get(m.myKey)
	// }
	m.myKey = fmt.Sprintf("%s%x", m.pfx, s.Lease())
	cmp := v3.Compare(v3.CreateRevision(m.myKey), "=", 0)
	put := v3.OpPut(m.myKey, "", v3.WithLease(s.Lease()))
	get := v3.OpGet(m.myKey)
	getOwner := v3.OpGet(m.pfx, v3.WithFirstCreate()...)
	resp, err := client.Txn(ctx).If(cmp).Then(put, getOwner).Else(get, getOwner).Commit()
	if err != nil {
		return err
	}
 	// 此锁是否为自己获得
	m.myRev = resp.Header.Revision
	if !resp.Succeeded {
		m.myRev = resp.Responses[0].GetResponseRange().Kvs[0].CreateRevision
	}
	ownerKey := resp.Responses[1].GetResponseRange().Kvs
	if len(ownerKey) == 0 || ownerKey[0].CreateRevision == m.myRev {
		m.hdr = resp.Header
		return nil
	}
	// 阻塞直到获取到该锁
	hdr, werr := waitDeletes(ctx, client, m.pfx, m.myRev-1)
	if werr != nil {
		m.Unlock(client.Ctx())
	} else {
		m.hdr = hdr
	}
	return werr
}
```

4. `func (m *Mutex) Unlock(ctx context.Context) error` 删除 key

# 


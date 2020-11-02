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

# 分布式锁


# grpc 负载均衡

服务端
```go
func main() {
	s := &grpclb.ServiceInfo{
		RegisterDir: "service",
		Name:        "greet",
		Version:     "v1",
		Endpoint:    "127.0.0.1:9001",
		TTL:         15,
	}
	lis, err := net.Listen("tcp", s.Endpoint)
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	registrar, err := grpclb.NewRegistrar()
	if err != nil {
		log.Fatalf("grpclb new registrar err: %s", err.Error())
	}
	go registrar.Register(s, clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	greetProto.RegisterGreetServiceServer(grpcServer, Handler{})
	grpcServer.Serve(lis)
}
```

客户端
```go
func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	b := grpc.RoundRobin(grpclb.NewEtcdResolver("service", "greet", cli))
	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithBalancer(b), grpc.WithBlock())
	if err != nil {
		log.Fatalf("err: %v", err.Error())
		return
	}
	defer conn.Close()
	client := greetProto.NewGreetServiceClient(conn)
	resp, err := client.Hi(context.TODO(), &greetProto.HiRequest{Name: "刘宁"})
	if err != nil {
		log.Fatalf("call hi err: %s", err.Error())
		return
	}
	fmt.Println(resp.Message)
}
```
package grpclb

type ServiceInfo struct {
	RegisterDir string `json:"register_dir"` // 服务注册前缀
	Name        string `json:"name"`         // 服务名称
	Version     string `json:"version"`      // 版本
	Endpoint    string `json:"endpoint"`     // 服务端口
	TTL         int64  `json:"ttl"`          // 服务健康检测间隔
}

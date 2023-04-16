package registry

import (
	"context"
	"errors"
)

// The Registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	// Register 注册服务
	Register(ctx context.Context, service Service, options ...RegisterOption) error
	// Deregister 取消注册服务
	Deregister(ctx context.Context, service Service) error
	// GetService 查询服务
	GetService(ctx context.Context, serviceName string) ([]Service, error)
	// ListServices 查询所有服务
	ListServices(ctx context.Context) ([]Service, error)
	// Watch 获取服务监听器
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

var (
	// ErrNotFound Not found error when GetService is called
	ErrNotFound = errors.New("service not found")
	// ErrWatcherStopped Watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
)

// Service 服务配置
type Service struct {
	Name      string            `json:"name"`      // 服务名称
	Version   string            `json:"version"`   // 服务版本号
	Metadata  map[string]string `json:"metadata"`  // 服务元数据，以键值对的形式保存附加信息
	Endpoints []Endpoint        `json:"endpoints"` // 服务端点列表
	Nodes     []Node            `json:"nodes"`     // 服务节点列表
}

// Endpoint 服务端点
type Endpoint struct {
	Name     string            `json:"name"`     // 端点名称
	Request  *Value            `json:"request"`  // 端点请求参数
	Response *Value            `json:"response"` // 端点响应参数
	Metadata map[string]string `json:"metadata"` // 端点元数据，以键值对的形式保存附加信息
}

// Value 服务参数
type Value struct {
	Name   string  `json:"name"`   // 参数名称
	Type   string  `json:"type"`   // 参数类型
	Values []Value `json:"values"` // 参数的值，如果参数是复杂类型，则该字段为参数列表
}

// Node 服务节点
type Node struct {
	Id       string            `json:"id"`       // 节点ID
	Address  string            `json:"address"`  // 节点的地址
	Metadata map[string]string `json:"metadata"` // 节点元数据，以键值对的形式保存附加信息
}

// +k8s:deepcopy-gen=package
package registry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrRegistry dsync errors.
	ErrRegistry = errors.New("registry")
	// ErrNotFound Not found error when IRegistry.GetService or IRegistry.GetServiceNode is called
	ErrNotFound = fmt.Errorf("%w: service not found", ErrRegistry)
	// ErrStoppedWatching Stopped watching error when watcher is stopped
	ErrStoppedWatching = fmt.Errorf("%w: stopped watching", ErrRegistry)
)

// The IRegistry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type IRegistry interface {
	// Register 注册服务
	Register(ctx context.Context, service Service, ttl time.Duration) error
	// Deregister 取消注册服务
	Deregister(ctx context.Context, service Service) error
	// GetServiceNode 查询服务节点
	GetServiceNode(ctx context.Context, serviceName, nodeId string) (*Service, error)
	// GetService 查询服务
	GetService(ctx context.Context, serviceName string) ([]Service, error)
	// ListServices 查询所有服务
	ListServices(ctx context.Context) ([]Service, error)
	// Watch 获取服务监听器
	Watch(ctx context.Context, pattern string) (IWatcher, error)
}

// Service 服务配置
// +k8s:deepcopy-gen=true
type Service struct {
	Name      string            `json:"name"`      // 服务名称
	Version   string            `json:"version"`   // 服务版本号
	Metadata  map[string]string `json:"metadata"`  // 服务元数据，以键值对的形式保存附加信息
	Endpoints []Endpoint        `json:"endpoints"` // 服务端点列表
	Nodes     []Node            `json:"nodes"`     // 服务节点列表
}

// Endpoint 服务端点
// +k8s:deepcopy-gen=true
type Endpoint struct {
	Name     string            `json:"name"`     // 端点名称
	Request  *Value            `json:"request"`  // 端点请求参数
	Response *Value            `json:"response"` // 端点响应参数
	Metadata map[string]string `json:"metadata"` // 端点元数据，以键值对的形式保存附加信息
}

// Value 服务参数
// +k8s:deepcopy-gen=true
type Value struct {
	Name   string  `json:"name"`   // 参数名称
	Type   string  `json:"type"`   // 参数类型
	Values []Value `json:"values"` // 参数的值，如果参数是复杂类型，则该字段为参数列表
}

// Node 服务节点
// +k8s:deepcopy-gen=true
type Node struct {
	Id       string            `json:"id"`       // 节点ID
	Address  string            `json:"address"`  // 节点的地址
	Metadata map[string]string `json:"metadata"` // 节点元数据，以键值对的形式保存附加信息
}

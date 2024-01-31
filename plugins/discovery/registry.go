// +k8s:deepcopy-gen=package
package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrRegistry registry errors.
	ErrRegistry = errors.New("registry")
	// ErrNotFound Not found error when IRegistry.GetService or IRegistry.GetServiceNode is called
	ErrNotFound = fmt.Errorf("%w: service not found", ErrRegistry)
	// ErrStoppedWatching Stopped watching error when watcher is stopped
	ErrStoppedWatching = fmt.Errorf("%w: stopped watching", ErrRegistry)
)

// Service 服务配置
// +k8s:deepcopy-gen=true
type Service struct {
	Name     string `json:"name"`     // 服务名称
	Nodes    []Node `json:"nodes"`    // 服务节点列表
	Revision int64  `json:"revision"` // 数据版本号
}

// Node 服务节点
// +k8s:deepcopy-gen=true
type Node struct {
	Id      string            `json:"id"`      // 节点ID
	Address string            `json:"address"` // 节点的地址
	Version string            `json:"version"` // 节点的服务版本号
	Meta    map[string]string `json:"meta"`    // 节点元数据，以键值对的形式保存附加信息
}

// The IRegistry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type IRegistry interface {
	// Register 注册服务
	Register(ctx context.Context, service *Service, ttl time.Duration) error
	// Deregister 取消注册服务
	Deregister(ctx context.Context, service *Service) error
	// GetServiceNode 查询服务节点
	GetServiceNode(ctx context.Context, serviceName, nodeId string) (*Service, error)
	// GetService 查询服务
	GetService(ctx context.Context, serviceName string) (*Service, error)
	// ListServices 查询所有服务
	ListServices(ctx context.Context) ([]Service, error)
	// Watch 获取服务监听器
	Watch(ctx context.Context, pattern string, revision ...int64) (IWatcher, error)
}

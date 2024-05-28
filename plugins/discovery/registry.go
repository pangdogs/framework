// +k8s:deepcopy-gen=package
package discovery

import (
	"context"
	"errors"
	"git.golaxy.org/core/utils/uid"
	"time"
)

var (
	// ErrNotFound Not found error when IRegistry.GetService or IRegistry.GetServiceNode is called
	ErrNotFound = errors.New("registry: service not found")
	// ErrTerminated Stopped watching error when watcher is stopped
	ErrTerminated = errors.New("registry: watching terminated")
)

// Service 服务配置
// +k8s:deepcopy-gen=true
type Service struct {
	Name     string `json:"name"`               // 服务名称
	Nodes    []Node `json:"nodes"`              // 服务节点列表
	Revision int64  `json:"revision,omitempty"` // 数据版本号
}

// Node 服务节点
// +k8s:deepcopy-gen=true
type Node struct {
	Id      uid.Id            `json:"id"`                // 节点ID
	Address string            `json:"address"`           // 节点的地址
	Version string            `json:"version,omitempty"` // 节点的服务版本号
	Meta    map[string]string `json:"meta,omitempty"`    // 节点元数据，以键值对的形式保存附加信息
}

// The IRegistry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type IRegistry interface {
	// Register 注册服务
	Register(ctx context.Context, service *Service, ttl time.Duration) error
	// Deregister 取消注册服务
	Deregister(ctx context.Context, service *Service) error
	// RefreshTTL 刷新所有服务TTL
	RefreshTTL(ctx context.Context) error
	// GetServiceNode 查询服务节点
	GetServiceNode(ctx context.Context, serviceName string, nodeId uid.Id) (*Service, error)
	// GetService 查询服务
	GetService(ctx context.Context, serviceName string) (*Service, error)
	// ListServices 查询所有服务
	ListServices(ctx context.Context) ([]Service, error)
	// Watch 监听服务变化
	Watch(ctx context.Context, pattern string, revision ...int64) (IWatcher, error)
}

/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

//go:generate go run k8s.io/code-generator/cmd/deepcopy-gen .
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

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

//go:generate stringer -type EventType
package discovery

import (
	"context"
	"errors"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/uid"
)

var (
	// ErrDuplicateRegistration 重复注册信息
	ErrDuplicateRegistration = errors.New("registry: duplicate registration")
	// ErrRegistrationNotFound 注册信息不存在
	ErrRegistrationNotFound = errors.New("registry: registration not found")
)

// Service 服务信息
type Service struct {
	Name     string `json:"name"`               // 服务名称
	Nodes    []Node `json:"nodes"`              // 节点列表
	Revision int64  `json:"revision,omitempty"` // 查询时的全局数据版本号
}

// Node 服务节点
type Node struct {
	Id      uid.Id            `json:"id"`                // 节点ID
	Address string            `json:"address"`           // 节点的地址
	Version string            `json:"version,omitempty"` // 节点的服务版本号
	Meta    map[string]string `json:"meta,omitempty"`    // 节点元数据，以键值对的形式保存附加信息
}

// EventType 服务变化事件类型
type EventType int8

const (
	EventType_Create EventType = iota // 创建
	EventType_Delete                  // 删除
	EventType_Update                  // 更新
	EventType_Error                   // 错误
)

// Event 服务变化事件
type Event struct {
	Type    EventType // 事件类型
	Service *Service  // 服务信息
	Error   error     // 错误
}

type (
	// EventHandler 服务事件处理器
	EventHandler = generic.DelegateVoid1[Event]
)

// IRegistry 服务注册接口
type IRegistry interface {
	// RegisterNode 注册服务节点
	RegisterNode(ctx context.Context, serviceName string, node *Node, ttl time.Duration) (IRegistration, error)
	// Get 查询服务
	Get(ctx context.Context, serviceName string) (*Service, error)
	// GetNode 查询服务节点
	GetNode(ctx context.Context, serviceName string, nodeId uid.Id) (*Service, error)
	// List 查询所有服务
	List(ctx context.Context) ([]*Service, error)
	// WatchEvent 观察服务变化事件流
	WatchEvent(ctx context.Context, pattern string, revision ...int64) (<-chan Event, error)
	// WatchHandler 观察服务变化事件回调
	WatchHandler(ctx context.Context, pattern string, handler EventHandler, revision ...int64) (async.Future, error)
}

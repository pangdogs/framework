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
	// ErrDuplicateRegistration 表示同一服务节点已经注册。
	ErrDuplicateRegistration = errors.New("registry: duplicate registration")
	// ErrRegistrationNotFound 表示请求的服务或节点不存在。
	ErrRegistrationNotFound = errors.New("registry: registration not found")
)

// Service 是一次服务发现查询或变化事件中的服务快照。
type Service struct {
	Name     string `json:"name"`               // Name 是逻辑服务名称。
	Nodes    []Node `json:"nodes"`              // Nodes 是当前已发现的服务节点。
	Revision int64  `json:"revision,omitempty"` // Revision 是生成快照时的存储修订号。
}

// Node 描述一个可被发现的服务节点。
type Node struct {
	ID      uid.ID            `json:"id"`                // ID 是节点 ID。
	Address string            `json:"address"`           // Address 是节点对外地址。
	Version string            `json:"version,omitempty"` // Version 是节点服务版本。
	Meta    map[string]string `json:"meta,omitempty"`    // Meta 保存节点附加元数据。
}

// EventType 表示一次服务发现变化的类别。
type EventType int8

const (
	// EventType_Create 表示节点注册被创建。
	EventType_Create EventType = iota
	// EventType_Delete 表示节点注册被删除。
	EventType_Delete
	// EventType_Update 表示节点注册被更新。
	EventType_Update
	// EventType_Error 表示监听流程发生错误；错误内容保存在 Event.Error。
	EventType_Error
)

// Event 描述服务发现监听器观察到的一次变化或错误。
type Event struct {
	Type    EventType // Type 是事件类别。
	Service *Service  // Service 是变化后的服务节点快照；错误事件中可能为 nil。
	Error   error     // Error 仅在 Type 为 EventType_Error 时有值。
}

type (
	// EventHandler 处理一次服务发现变化。
	EventHandler = generic.DelegateVoid1[Event]
)

// IRegistry 定义服务节点注册、查询及变化监听能力。
type IRegistry interface {
	// RegisterNode 注册一个带 ttl 租约的服务节点，并返回租约控制句柄。
	RegisterNode(ctx context.Context, serviceName string, node *Node, ttl time.Duration) (IRegistration, error)
	// Get 返回 serviceName 当前全部节点的快照。
	Get(ctx context.Context, serviceName string) (*Service, error)
	// GetNode 返回 serviceName 下指定节点的快照。
	GetNode(ctx context.Context, serviceName string, nodeID uid.ID) (*Service, error)
	// List 返回当前全部服务及节点的快照。
	List(ctx context.Context) ([]*Service, error)
	// WatchEvent 从可选 revision 开始监听 pattern 对应服务的变化。
	// pattern 为空时监听全部服务；ctx 取消时返回通道会关闭。
	WatchEvent(ctx context.Context, pattern string, revision ...int64) (<-chan Event, error)
	// WatchHandler 从可选 revision 开始监听变化并调用 handler。
	// 返回的 Signal 在 ctx 取消或监听结束后完成。
	WatchHandler(ctx context.Context, pattern string, handler EventHandler, revision ...int64) (async.Signal, error)
}

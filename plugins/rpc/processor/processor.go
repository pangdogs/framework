package processor

import (
	"errors"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
)

var (
	ErrNoDeliverer                  = errors.New("rpc: no deliverer")                        // 没有匹配的投递器
	ErrTerminated                   = errors.New("rpc: deliverer terminated")                // 已终止投递
	ErrEntityNotFound               = errors.New("rpc: session routing to entity not found") // 找不到路由会话映射的实体
	ErrSessionNotFound              = errors.New("rpc: entity routing to session not found") // 找不到路由实体映射的会话
	ErrGroupNotFound                = errors.New("rpc: group not found")                     // 找不到分组
	ErrGroupChanIsFull              = errors.New("rpc: group send data channel is full")     // 分组发送数据的channel已满
	ErrDistEntityNotFound           = errors.New("rpc: distributed entity not found")        // 找不到分布式实体
	ErrDistEntityNodeNotFound       = errors.New("rpc: distributed entity node not found")   // 找不到分布式实体的服务节点
	ErrIncorrectDestAddress         = errors.New("rpc: incorrect destination Address")       // 错误的目的地址
	ErrPluginNotFound               = errors.New("rpc: plugin not found")                    // 找不到插件
	ErrMethodNotFound               = errors.New("rpc: method not found")                    // 找不到方法
	ErrComponentNotFound            = errors.New("rpc: component not found")                 // 找不到组件
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")     // 方法参数数量不匹配
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")      // 方法参数类型不匹配
	ErrPermissionDenied             = errors.New("rpc: permission denied")                   // 权限不足
)

// IDeliverer RPC投递器接口，用于将RPC投递至目标
type IDeliverer interface {
	// Match 是否匹配
	Match(ctx service.Context, dst, path string, oneWay bool) bool
	// Request 请求
	Request(ctx service.Context, dst, path string, args []any) runtime.AsyncRet
	// Notify 通知
	Notify(ctx service.Context, dst, path string, args []any) error
}

// IDispatcher RPC分发器接口，用于分发RPC请求与响应
type IDispatcher any

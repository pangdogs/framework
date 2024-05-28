package processor

import (
	"errors"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/plugins/rpcstack"
)

var (
	ErrUndeliverable                = errors.New("rpc: undeliverable")                     // 无法投递
	ErrTerminated                   = errors.New("rpc: processor terminated")              // 已终止处理
	ErrEntityNotFound               = errors.New("rpc: routing to entity not found")       // 找不到路由会话映射的实体
	ErrSessionNotFound              = errors.New("rpc: routing to session not found")      // 找不到路由实体映射的会话
	ErrGroupNotFound                = errors.New("rpc: group not found")                   // 找不到分组
	ErrGroupChanIsFull              = errors.New("rpc: group send data channel is full")   // 分组发送数据的channel已满
	ErrDistEntityNotFound           = errors.New("rpc: distributed entity not found")      // 找不到分布式实体
	ErrDistEntityNodeNotFound       = errors.New("rpc: distributed entity node not found") // 找不到分布式实体的服务节点
	ErrIncorrectDestAddress         = errors.New("rpc: incorrect destination Address")     // 错误的目的地址
	ErrPluginNotFound               = errors.New("rpc: plugin not found")                  // 找不到插件
	ErrMethodNotFound               = errors.New("rpc: method not found")                  // 找不到方法
	ErrComponentNotFound            = errors.New("rpc: component not found")               // 找不到组件
	ErrMethodParameterCountMismatch = errors.New("rpc: method parameter count mismatch")   // 方法参数数量不匹配
	ErrMethodParameterTypeMismatch  = errors.New("rpc: method parameter type mismatch")    // 方法参数类型不匹配
	ErrPermissionDenied             = errors.New("rpc: permission denied")                 // 权限不足
)

// IDeliverer RPC投递器接口
type IDeliverer interface {
	// Match 是否匹配
	Match(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, oneWay bool) bool
	// Request 请求
	Request(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) async.AsyncRet
	// Notify 通知
	Notify(ctx service.Context, dst string, callChain rpcstack.CallChain, path string, args []any) error
}

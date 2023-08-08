//go:generate stringer -type SessionState
package gate

import (
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/golaxy/service"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
)

// SessionState 客户端会话状态
type SessionState int32

const (
	SessionState_Birth     SessionState = iota // 出生
	SessionState_Handshake                     // 与客户端握手中
	SessionState_Confirmed                     // 已确认客户端连接
	SessionState_Active                        // 客户端活跃
	SessionState_Inactive                      // 客户端不活跃，等待重连恢复中
	SessionState_Death                         // 已过期
)

// RecvData 接收的数据
type RecvData struct {
	Data      []byte // 数据
	Sequenced bool   // 是否有时序
	Error     error  // 错误信息
}

// RecvEvent 接收的自定义事件
type RecvEvent struct {
	Event protocol.Event[transport.Msg] // 消息事件
	Error error                         // 错误信息
}

// Session 客户端会话
type Session interface {
	context.Context
	fmt.Stringer
	// GetContext 获取服务上下文
	GetContext() service.Context
	// GetId 获取会话Id
	GetId() string
	// GetToken 获取token
	GetToken() string
	// GetState 获取会话状态
	GetState() SessionState
	// GetGroups 获取所属的会话组Id
	GetGroups() []string
	// GetListenAddr 获取监听地址
	GetListenAddr() net.Addr
	// GetClientAddr 获取客户端地址
	GetClientAddr() net.Addr
	// SendData 发送数据
	SendData(data []byte, sequenced bool) error
	// SendEvent 发送自定义事件
	SendEvent(event protocol.Event[transport.Msg]) error
	// RecvDataChan 接收数据的chan
	RecvDataChan() <-chan RecvData
	// RecvEventChan 接收自定义事件的chan
	RecvEventChan() <-chan RecvEvent
	// Close 关闭
	Close(err error)
}

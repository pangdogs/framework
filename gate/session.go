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
	// GetLocalAddr 获取本地地址
	GetLocalAddr() net.Addr
	// GetRemoteAddr 获取对端地址
	GetRemoteAddr() net.Addr
	// SendData 发送数据
	SendData(data []byte) error
	// SendEvent 发送自定义事件
	SendEvent(event protocol.Event[transport.Msg]) error
	// SendDataChan 发送数据的channel
	SendDataChan() chan<- []byte
	// RecvDataChan 接收数据的channel
	RecvDataChan() <-chan []byte
	// SendEventChan 发送自定义事件的channel
	SendEventChan() chan<- protocol.Event[transport.Msg]
	// RecvEventChan 接收自定义事件的channel
	RecvEventChan() <-chan protocol.Event[transport.Msg]
	// Close 关闭
	Close(err error)
}

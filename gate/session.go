package gate

import (
	"fmt"
	"io"
	"net"
)

// SessionState 客户端会话状态
type SessionState int32

const (
	SessionState_Handshake SessionState = iota // 握手中
	SessionState_Active                        // 活跃，客户端已连接
	SessionState_Inactive                      // 不活跃，客户端已断连，等待重连恢复中
)

// Session 客户端会话
type Session interface {
	fmt.Stringer
	io.Reader
	io.WriterTo
	io.Writer
	io.ReaderFrom
	// GetId 获取会话Id
	GetId() string
	// GetGroups 获取所属的会话组Id
	GetGroups() []string
	// GetToken 获取Token
	GetToken() string
	// GetListenAddr 获取监听地址
	GetListenAddr() net.Addr
	// GetClientAddr 获取客户端地址
	GetClientAddr() net.Addr
	// GetState 获取会话状态
	GetState() SessionState
	// Close 关闭连接
	Close()
}

package tcp

import (
	"kit.golaxy.org/plugins/gate"
	"net"
)

type _TcpSession struct {
	gate  *_TcpGate
	conn  net.Conn
	state gate.SessionState
}

// GetId 获取会话Id
func (s *_TcpSession) GetId() string {
	s.conn.
}

// GetGroups 获取所属的会话组Id
func (s *_TcpSession) GetGroups() []string {

}

// GetToken 获取Token
func (s *_TcpSession) GetToken() string {

}

// GetListenAddr 获取监听地址
func (s *_TcpSession) GetListenAddr() net.Addr {

}

// GetClientAddr 获取客户端地址
func (s *_TcpSession) GetClientAddr() net.Addr {

}

// GetState 获取会话状态
func (s *_TcpSession) GetState() gate.SessionState {

}

// Close 关闭连接
func (s *_TcpSession) Close() {

}

// Read 读取数据
func (s *_TcpSession) Read(b []byte) (int, error) {

}

// Write 写入数据
func (s *_TcpSession) Write(b []byte) (int, error) {

}

func (s *_TcpSession) String() string {

}

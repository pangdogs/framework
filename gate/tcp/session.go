package tcp

import (
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"sync"
)

func newTcpSession(tcpGate *_TcpGate) *_TcpSession {
	session := &_TcpSession{
		gate:  tcpGate,
		id:    ksuid.New().String(),
		state: gate.SessionState_Handshake,
	}

	session.dispatcher.Transceiver = &session.transceiver
	session.dispatcher.Add(&session.trans)
	session.dispatcher.Add(&session.ctrl)

	session.trans.Transceiver = &session.transceiver
	session.ctrl.Transceiver = &session.transceiver

	return session
}

type _TcpSession struct {
	gate        *_TcpGate
	ctx         context.Context
	id          string
	state       gate.SessionState
	token       string
	groups      []string
	transceiver protocol.Transceiver
	dispatcher  protocol.EventDispatcher
	trans       protocol.TransProtocol
	ctrl        protocol.CtrlProtocol
	sync.Mutex
}

func (s *_TcpSession) String() string {
	return fmt.Sprintf("{Id:%s State:%d}", s.GetId(), s.GetState())
}

// GetId 获取会话Id
func (s *_TcpSession) GetId() string {
	return s.id
}

// GetState 获取会话状态
func (s *_TcpSession) GetState() gate.SessionState {
	s.Lock()
	defer s.Unlock()

	return s.state
}

// GetToken 获取token
func (s *_TcpSession) GetToken() string {
	return s.token
}

// GetGroups 获取所属的会话组Id
func (s *_TcpSession) GetGroups() []string {
	s.Lock()
	defer s.Unlock()

	groups := make([]string, len(s.groups))
	copy(groups, s.groups)

	return groups
}

// GetListenAddr 获取监听地址
func (s *_TcpSession) GetListenAddr() net.Addr {
	s.Lock()
	defer s.Unlock()

	return s.transceiver.Conn.LocalAddr()
}

// GetClientAddr 获取客户端地址
func (s *_TcpSession) GetClientAddr() net.Addr {
	s.Lock()
	defer s.Unlock()

	return s.transceiver.Conn.RemoteAddr()
}

// Close 关闭连接
func (s *_TcpSession) Close(err error) {
	s.Lock()
	defer s.Unlock()

	if err != nil {
		s.ctrl.SendRst(err)
	}
	s.transceiver.Conn.Close()
}

func (s *_TcpSession) Init(transceiver protocol.Transceiver, token string) {
	s.transceiver = transceiver
	s.token = token
}

func (s *_TcpSession) Renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	s.Lock()
	defer s.Unlock()

	// 切换连接
	s.transceiver.Conn.Close()
	s.transceiver.Conn = conn

	// 同步对端时序
	if !s.transceiver.SequencedBuff.Synchronization(remoteRecvSeq) {
		return 0, 0, errors.New("sequenced buff synchronization failed")
	}

	return s.transceiver.SequencedBuff.SendSeq, s.transceiver.SequencedBuff.RecvSeq, nil
}

func (s *_TcpSession) Run() {

}

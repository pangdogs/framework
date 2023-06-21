package tcp

import (
	"crypto/rand"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
)

// RangeSessions 遍历所有会话
func (g *_TcpGate) RangeSessions(fun func(session gate.Session) bool) {

}

// CountSessions 统计所有会话数量
func (g *_TcpGate) CountSessions() int {

}

func (g *_TcpGate) newSession(conn net.Conn) {
	defer func() {
		if info := recover(); info != nil {
			logger.Errorf(g.ctx, "new session failed, listener %q client address %q, %s", conn.LocalAddr(), conn.RemoteAddr(), info)
			conn.Close()
		}
	}()

	// 握手
	err := g.handshake(conn)
	if err != nil {
		panic(err)
	}

}

func (g *_TcpGate) handshake(conn net.Conn) error {
	handshake := &protocol.HandshakeProtocol{
		Conn:    conn,
		Encoder: &codec.Encoder{},
		Decoder: &codec.Decoder{MsgCreator: g.options.MsgCreator},
		Timeout: g.options.Timeout,
	}

	cs := g.options.CipherSuite
	cm := g.options.CompressionMethod
	var sessionId []byte

	err := handshake.ServerHello(func(cliHello protocol.Event[*transport.MsgHello]) (protocol.Event[*transport.MsgHello], error) {
		// 检查协议版本
		if cliHello.Msg.Version != transport.Version_V1_0 {
			return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{Code: transport.Code_VersionError}
		}

		// 检测会话是否存在，存在走断连恢复流程
		if len(cliHello.Msg.SessionId) > 0 {

		} else {
			sessionId = rand.Read(rand.)
		}

		// 检测是否使用客户端要求的加密与密码学套件
		if g.options.AgreeCliProposal {
			cs = cliHello.Msg.CipherSuite
			cm = cliHello.Msg.CompressionMethod
		}

		// 随机数，用于秘钥交换
		n, err := rand.Prime(rand.Reader, 256)
		if err != nil {

		}

		// 返回
		servHello := protocol.Event[*transport.MsgHello]{
			Flags: transport.Flags(transport.Flag_HelloDone),
			Msg: &transport.MsgHello{
				Version:           transport.Version_V1_0,
				SessionId:         nil,
				Random:            n.Bytes(),
				CipherSuite:       transport.CipherSuite{},
				CompressionMethod: 0,
				Extensions:        nil,
			},
		}

		return servHello, nil
	})
	if err != nil {
		return err
	}

	// 选择秘钥交换函数
	switch cs.SecretKeyExchangeMethod {
	case transport.SecretKeyExchangeMethod_None:
		break
	case transport.MsgId_ECDHESecretKeyExchange:

	default:

	}


	return nil
}

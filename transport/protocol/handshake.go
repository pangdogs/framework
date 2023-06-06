package tcp

import (
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"math/rand"
	"net"
)

func NewHandshake(conn net.Conn) *Handshake {
	return &Handshake{
		conn: conn,
	}
}

type Handshake struct {
	conn     net.Conn
	cliHello *transport.MsgHello
}

func (h *Handshake) Hello() error {
	cliMp := transport.MsgPacket{}
	if _, err := cliMp.ReadFrom(h.conn); err != nil {
		return err
	}

	if cliMp.Msg.MsgId() != transport.MsgId_Hello {
		return errors.New()
	}

	if cliHello.Version != transport.Version_V1_0 {
		return fmt.Errorf("version 0x%x not supported", cliHello.Version)
	}

	if len(cliHello.SessionId) > 0 {

	}

	var flags transport.Flags
	flags.Set(transport.Flag_HelloDone, true)

	servHello := &transport.MsgHello{
		Version:           transport.Version_V1_0,
		SessionId:         nil,
		Random:            rand.Int63(),
		CipherSuite:       transport.CipherSuite{},
		CompressionMethod: 0,
		Extensions:        nil,
	}

	if err := h.sendMsg(flags, servHello); err != nil {
		return err
	}

	h.cliHello = cliHello

	return nil
}

func (h *Handshake) SecretKeyExchange() error {

}

func (h *Handshake) Auth() error {

}

func (h *Handshake) Create() error {

}

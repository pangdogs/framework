package gtp_client

import (
	"bytes"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"time"
)

// init 初始化
func (c *Client) init(transceiver *protocol.Transceiver, sessionId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 初始化消息收发器
	c.transceiver.Conn = transceiver.Conn
	c.transceiver.Encoder = transceiver.Encoder
	c.transceiver.Decoder = transceiver.Decoder
	c.transceiver.Timeout = transceiver.Timeout
	c.transceiver.SequencedBuff.Reset(transceiver.SequencedBuff.SendSeq, transceiver.SequencedBuff.RecvSeq, transceiver.SequencedBuff.Cap)

	// 初始化会话Id
	c.sessionId = sessionId
}

// renew 刷新
func (c *Client) renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 刷新链路
	sendSeq, recvSeq, err = c.transceiver.Renew(conn, remoteRecvSeq)
	if err != nil {
		return 0, 0, err
	}

	return c.transceiver.SequencedBuff.SendSeq, c.transceiver.SequencedBuff.RecvSeq, nil
}

// pauseIO 暂停收发消息
func (c *Client) pauseIO() {
	c.transceiver.Pause()
}

// continueIO 继续收发消息
func (c *Client) continueIO() {
	c.transceiver.Continue()
}

// run 运行
func (c *Client) run() {
	defer func() {
		if panicErr := util.Panic2Err(recover()); panicErr != nil {
			c.logger.Errorf("client %q panicked, %s", c.GetSessionId(), fmt.Errorf("panicked: %w", panicErr))
		}
		c.transceiver.Conn.Close()
	}()

	active := true
	pinged := false
	var timeout time.Time

	// 启动发送数据的线程
	if c.sendDataChan != nil {
		go func() {
			for {
				select {
				case sd := <-c.sendDataChan:
					if err := c.SendData(sd.Data, sd.Sequenced); err != nil {
						c.logger.Errorf("client %q fetch data from the send data channel for sending failed, %s", c.GetSessionId(), err)
					}
				case <-c.Done():
					return
				}
			}
		}()
	}

	// 启动发送自定义事件的线程
	if c.sendEventChan != nil {
		go func() {
			for {
				select {
				case e := <-c.sendEventChan:
					if err := c.SendEvent(e); err != nil {
						c.logger.Errorf("client %q fetch event from the send event channel for sending failed, %s", c.GetSessionId(), err)
					}
				case <-c.Done():
					return
				}
			}
		}()
	}

	for {
		// 非活跃状态，检测超时时间
		if !active {
			if time.Now().After(timeout) {
				c.cancel()
			}
		}

		// 检测连接是否已关闭
		select {
		case <-c.Done():
			return
		default:
		}

		// 分发消息事件
		if err := c.dispatcher.Dispatching(); err != nil {
			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, protocol.ErrTimeout) {
				if !pinged {
					c.ctrl.SendPing()
					pinged = true
				} else {
					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整连接状态不活跃
					if active {
						active = false
						timeout = time.Now().Add(c.options.InactiveTimeout)
					}
				}
				continue
			}

			// 其他网络io类错误，调整连接状态不活跃
			if errors.Is(err, protocol.ErrNetIO) {
				if active {
					active = false
					timeout = time.Now().Add(c.options.InactiveTimeout)
				}
				continue
			}

			c.logger.Debugf("client %q dispatching event failed, %s", c.GetSessionId(), err)
		}

		// 没有错误，或非网络io类错误，重置ping状态，调整连接状态活跃
		pinged = false
		active = true
	}
}

// eventHandler 接收自定义事件的处理器
func (c *Client) eventHandler(event protocol.Event[transport.Msg]) error {
	if c.recvEventChan != nil {
		select {
		case c.recvEventChan <- gate.RecvEvent{Event: event.Clone()}:
		default:
			c.logger.Errorf("client %q receive event channel is full", c.GetSessionId())
		}
	}

	for i := range c.options.RecvEventHandlers {
		handler := c.options.RecvEventHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(c, event) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

// payloadHandler Payload消息事件处理器
func (c *Client) payloadHandler(event protocol.Event[*transport.MsgPayload]) error {
	if c.recvDataChan != nil {
		select {
		case c.recvDataChan <- gate.RecvData{
			Data:      bytes.Clone(event.Msg.Data),
			Sequenced: event.Flags.Is(transport.Flag_Sequenced),
		}:
		default:
			c.logger.Errorf("client %q receive data channel is full", c.GetSessionId())
		}
	}

	for i := range c.options.RecvDataHandlers {
		handler := c.options.RecvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(c, event.Msg.Data, event.Flags.Is(transport.Flag_Sequenced)) })
		if err == nil || !errors.As(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

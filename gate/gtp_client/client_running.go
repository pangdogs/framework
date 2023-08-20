package gtp_client

import (
	"bytes"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/internal"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/protocol"
	"net"
	"time"
)

// init 初始化
func (c *Client) init(conn net.Conn, encoder codec.IEncoder, decoder codec.IDecoder, remoteSendSeq, remoteRecvSeq uint32, sessionId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 初始化消息收发器
	c.transceiver.Conn = conn
	c.transceiver.Encoder = encoder
	c.transceiver.Decoder = decoder
	c.transceiver.Timeout = c.options.IOTimeout

	buff := &protocol.SequencedBuffer{}
	buff.Reset(remoteRecvSeq, remoteSendSeq, c.options.IOBufferCap)

	c.transceiver.Buffer = buff

	// 初始化刷新通知channel
	c.renewChan = make(chan struct{}, 10)

	// 初始化会话Id
	c.sessionId = sessionId

	// 初始化channel
	if c.options.SendDataChanSize > 0 {
		c.sendDataChan = make(chan []byte, c.options.SendDataChanSize)
	}
	if c.options.RecvDataChanSize > 0 {
		c.recvDataChan = make(chan []byte, c.options.RecvDataChanSize)
	}
	if c.options.SendEventSize > 0 {
		c.sendEventChan = make(chan protocol.Event[transport.Msg], c.options.SendEventSize)
	}
	if c.options.RecvEventSize > 0 {
		c.recvEventChan = make(chan protocol.Event[transport.Msg], c.options.RecvEventSize)
	}
}

// renew 刷新
func (c *Client) renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	c.mutex.Lock()

	defer func() {
		c.mutex.Unlock()
		c.renewChan <- struct{}{}
	}()

	// 刷新链路
	return c.transceiver.Renew(conn, remoteRecvSeq)
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
		if c.transceiver.Conn != nil {
			c.transceiver.Conn.Close()
		}
		c.transceiver.Clean()
	}()

	active := true
	pinged := false
	var timeout time.Time

	// 启动发送数据的线程
	if c.sendDataChan != nil {
		go func() {
			for {
				select {
				case data := <-c.sendDataChan:
					if err := c.SendData(data); err != nil {
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
				case event := <-c.sendEventChan:
					if err := c.SendEvent(event); err != nil {
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

				func() {
					timer := time.NewTimer(10 * time.Second)
					defer timer.Stop()

					select {
					case <-timer.C:
						return
					case <-c.renewChan:
						return
					case <-c.Done():
						return
					}
				}()

				continue
			}

			c.logger.Debugf("client %q dispatching event failed, %s", c.GetSessionId(), err)
			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false

		// 调整连接状态活跃
		if !active {
			active = true

			// 发送缓存的消息
			protocol.Retry{
				Transceiver: &c.transceiver,
				Times:       c.options.IORetryTimes,
			}.Send(c.transceiver.Resend())
		}
	}
}

// eventHandler 接收自定义事件的处理器
func (c *Client) eventHandler(event protocol.Event[transport.Msg]) error {
	if c.recvEventChan != nil {
		select {
		case c.recvEventChan <- event.Clone():
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
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

// payloadHandler Payload消息事件处理器
func (c *Client) payloadHandler(event protocol.Event[*transport.MsgPayload]) error {
	if c.recvDataChan != nil {
		select {
		case c.recvDataChan <- bytes.Clone(event.Msg.Data):
		default:
			c.logger.Errorf("client %q receive data channel is full", c.GetSessionId())
		}
	}

	for i := range c.options.RecvDataHandlers {
		handler := c.options.RecvDataHandlers[i]
		if handler == nil {
			continue
		}
		err := internal.Call(func() error { return handler(c, event.Msg.Data) })
		if err == nil || !errors.Is(err, protocol.ErrUnexpectedMsg) {
			return err
		}
	}

	return protocol.ErrUnexpectedMsg
}

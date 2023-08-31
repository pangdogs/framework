package gtp_client

import (
	"bytes"
	"errors"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/internal"
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

	buff := &transport.SequencedBuffer{}
	buff.Reset(remoteRecvSeq, remoteSendSeq, c.options.IOBufferCap)

	c.transceiver.Buffer = buff

	// 初始化刷新通知channel
	c.renewChan = make(chan struct{}, 1)

	// 初始化自动重连channel
	if c.options.AutoReconnect {
		c.reconnectChan = make(chan struct{}, 1)
	}

	// 初始化会话Id
	c.sessionId = sessionId
}

// renew 刷新
func (c *Client) renew(conn net.Conn, remoteRecvSeq uint32) (sendSeq, recvSeq uint32, err error) {
	c.mutex.Lock()

	defer func() {
		c.mutex.Unlock()

		select {
		case c.renewChan <- struct{}{}:
		default:
		}
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
		c.logger.Debugf("client %q shutdown, conn %q -> %q", c.GetSessionId(), c.GetLocalAddr(), c.GetRemoteAddr())
	}()

	c.logger.Debugf("client %q started, conn %q -> %q", c.GetSessionId(), c.GetLocalAddr(), c.GetRemoteAddr())

	// 启动发送数据的线程
	if c.options.SendDataChan != nil {
		go func() {
			for {
				select {
				case data := <-c.options.SendDataChan:
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
	if c.options.SendEventChan != nil {
		go func() {
			for {
				select {
				case event := <-c.options.SendEventChan:
					if err := c.SendEvent(event); err != nil {
						c.logger.Errorf("client %q fetch event from the send event channel for sending failed, %s", c.GetSessionId(), err)
					}
				case <-c.Done():
					return
				}
			}
		}()
	}

	active := true
	pinged := false
	var timeout time.Time

	// 启动自动重连线程
	if c.options.AutoReconnect {
		go func() {
			for {
				select {
				case <-c.Done():
					return
				case <-c.reconnectChan:
					c.reconnect()
				}
			}
		}()
	}

	// 修改活跃状态
	changeActive := func(b bool) {
		active = b

		if !active && c.options.AutoReconnect {
			select {
			case c.reconnectChan <- struct{}{}:
			default:
			}
		}
	}

	for {
		// 非活跃状态，未开启自动重连，检测超时时间
		if !active && !c.options.AutoReconnect {
			if time.Now().After(timeout) {
				// 超时关闭连接
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
		if err := c.eventDispatcher.Dispatching(); err != nil {
			c.logger.Debugf("client %q dispatching event failed, %s", c.GetSessionId(), err)

			// 网络io超时，触发心跳检测，向对方发送ping
			if errors.Is(err, transport.ErrTimeout) {
				if !pinged {
					c.logger.Debugf("client %q send ping", c.GetSessionId())

					c.ctrl.SendPing()
					pinged = true
				} else {
					c.logger.Debugf("client %q no pong received", c.GetSessionId())

					// 未收到对方回复pong或其他消息事件，再次网络io超时，调整连接状态不活跃
					changeActive(false)
				}
				continue
			}

			// 其他网络io类错误，调整连接状态不活跃
			if errors.Is(err, transport.ErrNetIO) {
				changeActive(false)

				func() {
					timer := time.NewTimer(10 * time.Second)
					defer timer.Stop()

					select {
					case <-timer.C:
						return
					case <-c.renewChan:
						// 发送缓存的消息
						transport.Retry{
							Transceiver: &c.transceiver,
							Times:       c.options.IORetryTimes,
						}.Send(c.transceiver.Resend())
						// 重置ping状态
						pinged = false
						return
					case <-c.Done():
						return
					}
				}()

				c.logger.Debugf("client %q retry dispatching event, conn %q -> %q", c.GetSessionId(), c.GetLocalAddr(), c.GetRemoteAddr())
				continue
			}

			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false
		// 调整连接状态活跃
		changeActive(true)
	}
}

// reconnect 重连
func (c *Client) reconnect() {
	defer func() {
		// 释放自动重连channel
		for {
			select {
			case <-c.reconnectChan:
			default:
				return
			}
		}
	}()

	// 尝试重连
	for i := 0; c.options.AutoReconnectRetryTimes <= 0 || i < c.options.AutoReconnectRetryTimes; i++ {
		select {
		case <-c.Done():
			return
		default:
		}

		if err := Reonnect(c); err != nil {
			c.logger.Errorf("client %q auto reconnect failed, retry %d times, %s", c.GetSessionId(), i+1, err)

			// 服务端返回rst拒绝连接，刷新链路失败，这两种情况下不再重试，关闭客户端
			var rstErr *transport.RstError
			if errors.As(err, &rstErr) || errors.Is(err, transport.ErrRenewConn) {
				c.cancel()
				return
			}

			time.Sleep(c.options.AutoReconnectInterval)
			continue
		}

		c.logger.Debugf("client %q auto reconnect ok, retry %d times, conn %q -> %q", c.GetSessionId(), i+1, c.GetLocalAddr(), c.GetRemoteAddr())
		return
	}

	// 多次重连失败，关闭连接
	c.cancel()
}

// eventHandler 接收自定义事件的处理器
func (c *Client) eventHandler(event transport.Event[gtp.Msg]) error {
	if c.options.RecvEventChan != nil {
		select {
		case c.options.RecvEventChan <- event.Clone():
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
		if err != nil {
			c.logger.Errorf("client %q receive event handler error: %s", c.GetSessionId(), err)
		}
	}

	return transport.ErrUnexpectedMsg
}

// payloadHandler Payload消息事件处理器
func (c *Client) payloadHandler(event transport.Event[*gtp.MsgPayload]) error {
	if c.options.RecvDataChan != nil {
		select {
		case c.options.RecvDataChan <- bytes.Clone(event.Msg.Data):
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
		if err != nil {
			c.logger.Errorf("client %q receive data event handler error: %s", c.GetSessionId(), err)
		}
	}

	return nil
}

// heartbeatHandler Heartbeat消息事件处理器
func (c *Client) heartbeatHandler(event transport.Event[*gtp.MsgHeartbeat]) error {
	if event.Flags.Is(gtp.Flag_Ping) {
		c.logger.Debugf("client %q receive ping", c.GetSessionId())
	} else {
		c.logger.Debugf("client %q receive pong", c.GetSessionId())
	}
	return nil
}

// syncTimeHandler SyncTime消息事件处理器
func (c *Client) syncTimeHandler(event transport.Event[*gtp.MsgSyncTime]) error {
	if event.Flags.Is(gtp.Flag_RespTime) {
		c.logger.Debugf("client %q receive sync time, remote unix time: %d, local request unix time: %d",
			c.GetSessionId(), event.Msg.LocalUnixMilli, event.Msg.RemoteUnixMilli)

		respTime := &ResponseTime{
			RequestTime: time.UnixMilli(event.Msg.RemoteUnixMilli),
			LocalTime:   time.Now(),
			RemoteTime:  time.UnixMilli(event.Msg.LocalUnixMilli),
		}

		c.asyncDispatcher.Dispatching(event.Msg.ReqId, respTime, nil)
	}
	return nil
}

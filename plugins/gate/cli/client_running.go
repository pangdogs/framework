package cli

import (
	"bytes"
	"errors"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/util/binaryutil"
	"git.golaxy.org/framework/util/concurrent"
	"net"
	"time"
)

// init 初始化
func (c *Client) init(conn net.Conn, encoder codec.IEncoder, decoder codec.IDecoder, remoteSendSeq, remoteRecvSeq uint32, sessionId uid.Id) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 初始化消息收发器
	c.transceiver.Conn = conn
	c.transceiver.Encoder = encoder
	c.transceiver.Decoder = decoder
	c.transceiver.Timeout = c.options.IOTimeout
	c.transceiver.Synchronizer = transport.NewSequencedSynchronizer(remoteRecvSeq, remoteSendSeq, c.options.IOBufferCap)

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
	defer c.mutex.Unlock()

	// 刷新链路
	sendSeq, recvSeq, err = c.transceiver.Renew(conn, remoteRecvSeq)
	if err != nil {
		return
	}

	// 通知刷新
	select {
	case c.renewChan <- struct{}{}:
	default:
	}

	return
}

// pauseIO 暂停收发消息
func (c *Client) pauseIO() {
	c.transceiver.Pause()
}

// continueIO 继续收发消息
func (c *Client) continueIO() {
	c.transceiver.Continue()
}

// mainLoop 主线程
func (c *Client) mainLoop() {
	defer func() {
		c.terminate()

		if c.transceiver.Conn != nil {
			c.transceiver.Conn.Close()
		}
		c.transceiver.Clean()

		c.wg.Done()
		c.wg.Wait()
		close(c.closedChan)
	}()

	c.logger.Infof("client %q started, conn %q -> %q", c.GetSessionId(), c.GetLocalAddr(), c.GetRemoteAddr())

	// 启动发送数据的线程
	if c.options.SendDataChan != nil {
		go func() {
			defer func() {
				for bs := range c.options.SendDataChan {
					bs.Release()
				}
			}()
			for {
				select {
				case bs := <-c.options.SendDataChan:
					err := c.SendData(bs.Data())
					bs.Release()
					if err != nil {
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

	// 启动自动重连线程
	if c.options.AutoReconnect {
		go func() {
			for {
				select {
				case <-c.reconnectChan:
					c.reconnect()
				case <-c.Done():
					return
				}
			}
		}()
	}

	active := true
	pinged := false
	var timeout time.Time

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

loop:
	for {
		// 非活跃状态，未开启自动重连，检测超时时间
		if !active && !c.options.AutoReconnect {
			if time.Now().After(timeout) {
				c.terminate()
			}
		}

		// 检测连接是否已关闭
		select {
		case <-c.Done():
			break loop
		default:
		}

		// 分发消息事件
		if err := c.eventDispatcher.Dispatching(); err != nil {
			// 网络io错误
			if errors.Is(err, transport.ErrNetIO) {
				// 网络io超时，触发心跳检测，向对方发送ping
				if errors.Is(err, transport.ErrTimeout) {
					if !pinged {
						// 尝试ping对端
						c.logger.Debugf("client %q send ping", c.GetSessionId())
						c.ctrl.SendPing()
						pinged = true
					} else {
						// 未收到对方回复pong或其他消息事件，再次网络io超时，调整连接状态不活跃
						c.logger.Debugf("client %q no pong received", c.GetSessionId())
						changeActive(false)
					}
					continue
				}

				// 其他网络io类错误，调整连接状态不活跃
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

			c.logger.Errorf("client %q dispatching event failed, %s", c.GetSessionId(), err)
			continue
		}

		// 没有错误，或非网络io类错误，重置ping状态
		pinged = false
		// 调整连接状态活跃
		changeActive(true)
	}

	c.logger.Infof("client %q stopped, conn %q -> %q", c.GetSessionId(), c.GetLocalAddr(), c.GetRemoteAddr())
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

	c.logger.Infof("client %q auto reconnect started", c.GetSessionId())

	// 尝试重连
	for i := 0; c.options.AutoReconnectRetryTimes <= 0 || i < c.options.AutoReconnectRetryTimes; i++ {
		select {
		case <-c.Done():
			c.logger.Errorf("client %q auto reconnect aborted, client is closed", c.GetSessionId())
			return
		default:
		}

		if err := Reconnect(c); err != nil {
			c.logger.Errorf("client %q auto reconnect failed, retry %d times, %s", c.GetSessionId(), i+1, err)

			// 服务端返回rst拒绝连接，刷新链路失败，这两种情况下不再重试，关闭客户端
			var rstErr *transport.RstError
			if errors.As(err, &rstErr) || errors.Is(err, transport.ErrRenewConn) {
				c.logger.Errorf("client %q auto reconnect aborted, %s, close client", c.GetSessionId(), err)
				c.terminate()
				return
			}

			// 重连失败，暂停一会再试
			time.Sleep(c.options.AutoReconnectInterval)
			continue
		}

		c.logger.Infof("client %q auto reconnect success, retry %d times, conn %q -> %q", c.GetSessionId(), i+1, c.GetLocalAddr(), c.GetRemoteAddr())
		return
	}

	// 多次重连失败，关闭连接
	c.logger.Errorf("client %q auto reconnect unsuccessful, close client", c.GetSessionId())
	c.terminate()
}

// handleRecvEventChan 接收自定义事件并写入channel
func (c *Client) handleRecvEventChan(event transport.Event[gtp.Msg]) error {
	// 写入channel
	if c.options.RecvEventChan != nil {
		copied := event
		copied.Msg = event.Msg.Clone()

		select {
		case c.options.RecvEventChan <- copied:
		default:
			return errors.New("receive event channel is full")
		}
	}
	return nil
}

// handleEventProcess 接收自定义事件并回调
func (c *Client) handleEventProcess(event transport.Event[gtp.Msg]) error {
	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	// 回调监控器
	c.eventWatchers.AutoRLock(func(watchers *[]*_EventWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, event)
		}
	})

	// 回调处理器
	c.options.RecvEventHandler.Exec(interrupt, event)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// handleRecvDataChan 接收Payload消息数据并写入channel
func (c *Client) handleRecvDataChan(event transport.Event[gtp.MsgPayload]) error {
	// 写入channel
	if c.options.RecvDataChan != nil {
		bs := func() binaryutil.RecycleBytes {
			if c.options.RecvDataChanRecyclable {
				return binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Clone(event.Msg.Data))
			} else {
				return binaryutil.MakeNonRecycleBytes(bytes.Clone(event.Msg.Data))
			}
		}()

		select {
		case c.options.RecvDataChan <- bs:
		default:
			bs.Release()
			return errors.New("receive data channel is full")
		}
	}
	return nil
}

// handlePayloadProcess 接收Payload消息数据并回调
func (c *Client) handlePayloadProcess(event transport.Event[gtp.MsgPayload]) error {
	var errs []error

	interrupt := func(err, _ error) bool {
		if err != nil {
			errs = append(errs, err)
		}
		return false
	}

	// 回调监控器
	c.dataWatchers.AutoRLock(func(watchers *[]*_DataWatcher) {
		for i := range *watchers {
			(*watchers)[i].handler.Exec(interrupt, event.Msg.Data)
		}
	})

	// 回调处理器
	c.options.RecvDataHandler.Exec(interrupt, event.Msg.Data)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// handleHeartbeat Heartbeat消息事件处理器
func (c *Client) handleHeartbeat(event transport.Event[gtp.MsgHeartbeat]) error {
	if event.Flags.Is(gtp.Flag_Ping) {
		c.logger.Debugf("client %q receive ping", c.GetSessionId())
	} else {
		c.logger.Debugf("client %q receive pong", c.GetSessionId())
	}
	return nil
}

// handleSyncTime SyncTime消息事件处理器
func (c *Client) handleSyncTime(event transport.Event[gtp.MsgSyncTime]) error {
	if event.Flags.Is(gtp.Flag_RespTime) {
		respTime := &ResponseTime{
			RequestTime: time.UnixMilli(event.Msg.RemoteUnixMilli),
			LocalTime:   time.Now(),
			RemoteTime:  time.UnixMilli(event.Msg.LocalUnixMilli),
		}
		return c.futures.Resolve(event.Msg.CorrId, concurrent.MakeRet[any](respTime, nil))
	}
	return nil
}

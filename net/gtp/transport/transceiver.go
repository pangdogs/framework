/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var (
	ErrTrans            = errors.New("gtp-trans")                // 传输错误
	ErrNetIO            = fmt.Errorf("%w: net i/o", ErrTrans)    // 网络io错误
	ErrRenew            = fmt.Errorf("%w: renew conn", ErrTrans) // 刷新链路错误
	ErrDeadlineExceeded = os.ErrDeadlineExceeded                 // 网络io超时
	ErrClosed           = os.ErrClosed                           // 网络链路已关闭
	ErrShortBuffer      = io.ErrShortBuffer                      // 缓冲区不足
	ErrShortWrite       = io.ErrShortWrite                       // 短写
	ErrUnexpectedEOF    = io.ErrUnexpectedEOF                    // 非预期的io结束
	EOF                 = io.EOF                                 // io结束
)

// Transceiver 消息事件收发器，线程安全
type Transceiver struct {
	Conn                 net.Conn       // 网络连接
	Encoder              *codec.Encoder // 消息包编码器
	Decoder              *codec.Decoder // 消息包解码器
	Timeout              time.Duration  // 网络io超时时间
	Synchronizer         ISynchronizer  // 同步器
	buffer               bytes.Buffer   // 接收消息缓存
	sendMutex, recvMutex sync.Mutex     // 发送与接收消息锁
}

// Send 发送消息
func (t *Transceiver) Send(e IEvent) error {
	if t.Conn == nil {
		return fmt.Errorf("%w: Conn is nil", ErrTrans)
	}

	if t.Encoder == nil {
		return fmt.Errorf("%w: Encoder is nil", ErrTrans)
	}

	if t.Synchronizer == nil {
		return fmt.Errorf("%w: Synchronizer is nil", ErrTrans)
	}

	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	// 写入同步器
	if err := t.writeToSynchronizer(e); err != nil {
		return err
	}

	// 设置链路超时时间
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("%w: set conn send timeout failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("%w: send msg-packet failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
	}

	return nil
}

// SendRst 发送Rst消息事件
func (t *Transceiver) SendRst(err error) error {
	// 包装错误信息
	var rstErr *RstError
	if !errors.As(err, &rstErr) {
		rstErr = &RstError{Code: gtp.Code_Reject}
		if err != nil {
			rstErr.Message = err.Error()
		}
	}
	return t.Send(rstErr.ToEvent().Interface())
}

// Resend 重新发送未完整发送的消息事件
func (t *Transceiver) Resend() error {
	if t.Conn == nil {
		return fmt.Errorf("%w: Conn is nil", ErrTrans)
	}

	if t.Encoder == nil {
		return fmt.Errorf("%w: Encoder is nil", ErrTrans)
	}

	if t.Synchronizer == nil {
		return fmt.Errorf("%w: Synchronizer is nil", ErrTrans)
	}

	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	// 设置链路超时时间
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("%w: set conn resend timeout failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("%w: resend msg-packet failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
	}

	return nil
}

// Recv 接收消息事件
func (t *Transceiver) Recv(ctx context.Context) (IEvent, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if t.Conn == nil {
		return IEvent{}, fmt.Errorf("%w: Conn is nil", ErrTrans)
	}

	if t.Decoder == nil {
		return IEvent{}, fmt.Errorf("%w: Decoder is nil", ErrTrans)
	}

	if t.Synchronizer == nil {
		return IEvent{}, fmt.Errorf("%w: Synchronizer is nil", ErrTrans)
	}

	t.recvMutex.Lock()
	defer t.recvMutex.Unlock()

	var mpLen int
	var mpCache [bytes.MinRead]byte

	for {
		select {
		case <-ctx.Done():
			return IEvent{}, fmt.Errorf("%w: %w", ErrTrans, context.Canceled)
		default:
		}

		bufLen := t.buffer.Len()

		if bufLen > 0 && bufLen >= mpLen {
			// 解码消息
			mp, l, err := t.Decoder.Decode(t.buffer.Bytes(), t.Synchronizer)
			if err == nil {
				event := IEvent{
					Flags: mp.Head.Flags,
					Seq:   mp.Head.Seq,
					Ack:   mp.Head.Ack,
					Msg:   mp.Msg,
				}

				t.buffer.Next(l)
				t.Synchronizer.Ack(mp.Head.Ack)

				return event, nil
			}

			if !errors.Is(err, ErrShortBuffer) {
				t.buffer.Next(l)
				return IEvent{}, fmt.Errorf("%w: decode msg-packet failed, %w", ErrTrans, err)
			}

			// 消息长度
			mpLen = l
		}

		// 设置链路超时时间
		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return IEvent{}, fmt.Errorf("%w: set conn recv timeout failed, %w: %w", ErrTrans, ErrNetIO, err)
			}
		}

		for {
			// 从链路读取消息
			n, err := t.Conn.Read(mpCache[:])
			if err != nil {
				return IEvent{}, fmt.Errorf("%w: recv msg-packet failed, %w: %w", ErrTrans, ErrNetIO, err)
			}

			// 写入消息缓存
			if n > 0 {
				t.buffer.Write(mpCache[:n])
			}

			if mpLen <= 0 || t.buffer.Len() >= mpLen {
				break
			}
		}
	}
}

// Renew 刷新链路
func (t *Transceiver) Renew(conn net.Conn, remoteRecvSeq uint32) (sendReq, recvReq uint32, err error) {
	if conn == nil {
		return 0, 0, fmt.Errorf("%w: conn is nil", ErrRenew)
	}

	if t.Synchronizer == nil {
		return 0, 0, fmt.Errorf("%w: Synchronizer is nil", ErrRenew)
	}

	// 同步对端时序
	if err = t.Synchronizer.Synchronize(remoteRecvSeq); err != nil {
		return 0, 0, fmt.Errorf("%w: synchronize sequence failed, %s", ErrRenew, err)
	}

	// 切换连接
	if t.Conn != nil {
		t.Conn.Close()
	}
	t.Conn = conn

	// 清除缓存
	t.buffer.Reset()

	return t.Synchronizer.SendSeq(), t.Synchronizer.RecvSeq(), nil
}

// Pause 暂停收发消息
func (t *Transceiver) Pause() {
	t.sendMutex.Lock()
	t.recvMutex.Lock()
}

// Continue 继续收发消息
func (t *Transceiver) Continue() {
	t.recvMutex.Unlock()
	t.sendMutex.Unlock()
}

// GC GC
func (t *Transceiver) GC() {
	if t.Decoder != nil {
		t.Decoder.GC()
	}
}

// Clean 清理
func (t *Transceiver) Clean() {
	t.GC()

	if t.Synchronizer != nil {
		t.Synchronizer.Clean()
	}

	t.buffer.Reset()
}

func (t *Transceiver) writeToSynchronizer(e IEvent) error {
	// 编码消息
	buf, err := t.Encoder.Encode(e.Flags, e.Msg)
	if err != nil {
		return fmt.Errorf("%w: encode msg failed, %w", ErrTrans, err)
	}
	defer buf.Release()

	// 写入同步器
	if _, err = t.Synchronizer.Write(buf.Data()); err != nil {
		return fmt.Errorf("%w: write msg to synchronizer failed, %w", ErrTrans, err)
	}

	return nil
}

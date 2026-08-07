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
	"io"
	"net"
	"os"
	"sync"
	"time"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
)

var (
	// ErrTrans 是 GTP 传输错误的根错误。
	ErrTrans = errors.New("gtp-trans")
	// ErrNetIO 表示底层网络 I/O 失败。
	ErrNetIO = fmt.Errorf("%w: net i/o", ErrTrans)
	// ErrMigrateConn 表示迁移连接失败。
	ErrMigrateConn = fmt.Errorf("%w: migrate conn", ErrTrans)
	// ErrDeadlineExceeded 是网络 I/O 超时的标准错误别名。
	ErrDeadlineExceeded = os.ErrDeadlineExceeded
	// ErrClosed 是网络连接已关闭的标准错误别名。
	ErrClosed = os.ErrClosed
	// ErrShortBuffer 是缓冲区不足的标准错误别名。
	ErrShortBuffer = io.ErrShortBuffer
	// ErrShortWrite 是未完整写入的标准错误别名。
	ErrShortWrite = io.ErrShortWrite
	// ErrUnexpectedEOF 是数据意外结束的标准错误别名。
	ErrUnexpectedEOF = io.ErrUnexpectedEOF
	// EOF 是输入正常结束的标准错误别名。
	EOF = io.EOF
)

// Transceiver 在网络连接上收发 GTP 事件。
//
// Send、Resend 与 Migrate 的发送部分相互串行，Recv 与 Migrate 的接收部分相互串行，
// 因此可以由不同 goroutine 同时收发。调用方应在并发收发前完成字段配置，且不得并发调用 Dispose。
type Transceiver struct {
	Conn           net.Conn       // 当前网络连接。
	Encoder        *codec.Encoder // 消息包编码器。
	Decoder        *codec.Decoder // 消息包解码器。
	Timeout        time.Duration  // 单次网络 I/O 超时时间；小于等于零时不设置 deadline。
	Synchronizer   ISynchronizer  // 消息时序同步器。
	buffer         bytes.Buffer   // 尚未完成解码的接收数据。
	sendMu, recvMu sync.Mutex     // 分别串行化发送与接收操作。
}

// Send 编码并发送事件；发生短写或超时时，未确认数据仍保留在同步器中，可由 Resend 补发。
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

	t.sendMu.Lock()
	defer t.sendMu.Unlock()

	// 编码后的帧先进入同步器，网络失败时仍可补发。
	if err := t.writeToSynchronizer(e); err != nil {
		return err
	}

	// 每次实际写入前刷新 deadline，Timeout 为零时沿用连接现状。
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("%w: set conn send timeout failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// WriteTo 从当前补发位置继续写出所有缓存帧。
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("%w: send msg-packet failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
	}

	return nil
}

// SendRst 将 err 转换为链路重置事件并发送；非 RstError 使用拒绝码和原错误文本。
func (t *Transceiver) SendRst(err error) error {
	// 普通错误统一映射为 Reject，协议错误保留其 RST 代码。
	var rstErr *RstError
	if !errors.As(err, &rstErr) {
		rstErr = &RstError{Code: gtp.Code_Reject}
		if err != nil {
			rstErr.Message = err.Error()
		}
	}
	return t.Send(rstErr.ToEvent().Interface())
}

// Resend 重新发送同步器中尚未确认的缓存数据，不会重新编码事件。
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

	t.sendMu.Lock()
	defer t.sendMu.Unlock()

	// 补发也使用独立的完整写超时窗口。
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("%w: set conn resend timeout failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// 不重新编码，仅从同步器当前补发位置继续写出。
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("%w: resend msg-packet failed, cached: %d, %w: %w", ErrTrans, t.Synchronizer.Cached(), ErrNetIO, err)
	}

	return nil
}

// Recv 阻塞至收到一个完整事件、发生错误或在读取前观察到 ctx 取消。
// ctx 取消不能直接中断已经阻塞的 net.Conn.Read；需要及时退出时应同时配置 Timeout 或操作连接。
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

	t.recvMu.Lock()
	defer t.recvMu.Unlock()

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
			// mpLen 为零时 Decoder 会先探测包长；数据足够时再完整解码。
			mp, l, err := t.Decoder.Decode(t.buffer.Bytes(), t.Synchronizer)
			if err == nil {
				event := IEvent{
					Flags: mp.Head.Flags,
					Seq:   mp.Head.Seq,
					Ack:   mp.Head.Ack,
					Msg:   mp.Body,
				}

				t.buffer.Next(l)
				t.Synchronizer.Ack(mp.Head.Ack)

				return event, nil
			}

			if !errors.Is(err, ErrShortBuffer) {
				t.buffer.Next(l)
				return IEvent{}, fmt.Errorf("%w: decode msg-packet failed, %w", ErrTrans, err)
			}

			// 短缓冲错误同时返回完整包长，供后续读取判断。
			mpLen = l
		}

		// 每轮底层读取前刷新 deadline。
		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return IEvent{}, fmt.Errorf("%w: set conn recv timeout failed, %w: %w", ErrTrans, ErrNetIO, err)
			}
		}

		for {
			// 单次读取可能包含半包或多包，统一追加到接收缓存。
			n, err := t.Conn.Read(mpCache[:])
			if err != nil {
				return IEvent{}, fmt.Errorf("%w: recv msg-packet failed, %w: %w", ErrTrans, ErrNetIO, err)
			}

			if n > 0 {
				t.buffer.Write(mpCache[:n])
			}

			if mpLen <= 0 || t.buffer.Len() >= mpLen {
				break
			}
		}
	}
}

// Migrate 暂停收发，将同步器推进至 remoteRecvSeq 后切换连接。
// 旧连接会被关闭，未确认数据保留供后续 Resend 使用，接收缓存会被丢弃。
func (t *Transceiver) Migrate(conn net.Conn, remoteRecvSeq uint32) (sendReq, recvReq uint32, err error) {
	if conn == nil {
		return 0, 0, fmt.Errorf("%w: conn is nil", ErrMigrateConn)
	}

	if t.Synchronizer == nil {
		return 0, 0, fmt.Errorf("%w: Synchronizer is nil", ErrMigrateConn)
	}

	t.pause()
	defer t.resume()

	// 先按对端期望序号定位补发起点，失败时保留原连接。
	if err = t.Synchronizer.Synchronize(remoteRecvSeq); err != nil {
		return 0, 0, fmt.Errorf("%w: synchronize sequence failed, %s", ErrMigrateConn, err)
	}

	// 收发锁同时持有期间关闭旧连接并发布新连接。
	if t.Conn != nil {
		t.Conn.Close()
	}
	t.Conn = conn

	// 旧连接上的未解码接收字节不能跨连接复用。
	t.buffer.Reset()

	return t.Synchronizer.SendSeq(), t.Synchronizer.RecvSeq(), nil
}

// Dispose 释放同步器与接收缓存；它不会关闭 Conn，且不得与收发或迁移并发调用。
func (t *Transceiver) Dispose() {
	if t.Synchronizer != nil {
		t.Synchronizer.Dispose()
	}

	t.buffer.Reset()
}

func (t *Transceiver) writeToSynchronizer(e IEvent) error {
	// 编码器生成池化消息包，写入同步器后即可释放。
	buf, err := t.Encoder.Encode(e.Flags, e.Msg)
	if err != nil {
		return fmt.Errorf("%w: encode msg failed, %w", ErrTrans, err)
	}
	defer buf.Release()

	// 同步器会改写序号和确认号，并持有自己的帧副本。
	if _, err = t.Synchronizer.Write(buf.Payload()); err != nil {
		return fmt.Errorf("%w: write msg to synchronizer failed, %w", ErrTrans, err)
	}

	return nil
}

func (t *Transceiver) pause() {
	t.sendMu.Lock()
	t.recvMu.Lock()
}

func (t *Transceiver) resume() {
	t.sendMu.Unlock()
	t.recvMu.Unlock()
}

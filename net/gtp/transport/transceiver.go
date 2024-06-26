package transport

import (
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
	ErrRenewConn     = errors.New("gtp: renew conn")          // 刷新链路错误
	ErrUnexpectedSeq = errors.New("gtp: unexpected sequence") // 收到非预期的消息序号，表示序号不连续
	ErrDiscardSeq    = errors.New("gtp: discard sequence")    // 收到已过期的消息序号，表示次消息已收到过
	ErrNetIO         = errors.New("gtp: net i/o")             // 网络io错误
	ErrTimeout       = os.ErrDeadlineExceeded                 // 网络io超时
	ErrClosed        = os.ErrClosed                           // 网络链路已关闭
	ErrUnexpectedEOF = io.ErrUnexpectedEOF                    // 非预期的io结束
	EOF              = io.EOF                                 // io结束
)

// Transceiver 消息事件收发器，线程安全
type Transceiver struct {
	Conn                 net.Conn       // 网络连接
	Encoder              codec.IEncoder // 消息包编码器
	Decoder              codec.IDecoder // 消息包解码器
	Timeout              time.Duration  // 网络io超时时间
	Synchronizer         ISynchronizer  // 同步器缓存
	sendMutex, recvMutex sync.Mutex     // 发送与接收消息锁
}

// Send 发送消息
func (t *Transceiver) Send(e IEvent) error {
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("gtp: setting Conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("gtp: setting Encoder is nil")
	}

	if t.Synchronizer == nil {
		return errors.New("gtp: setting Synchronizer is nil")
	}

	// 编码消息
	if err := t.Encoder.EncodeWriter(t.Synchronizer, e.Flags, e.Msg); err != nil {
		return fmt.Errorf("gtp: encode msg failed, %w", err)
	}

	// 设置链路超时时间
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("gtp: set conn send timeout failed, cached: %d, %w: %w", t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("gtp: send msg-packet failed, cached: %d, %w: %w", t.Synchronizer.Cached(), ErrNetIO, err)
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
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("gtp: setting Conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("gtp: setting Encoder is nil")
	}

	if t.Synchronizer == nil {
		return errors.New("gtp: setting Synchronizer is nil")
	}

	// 设置链路超时时间
	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("gtp: set conn send timeout failed, cached: %d, %w: %w", t.Synchronizer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Synchronizer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("gtp: resend msg-packet failed, cached: %d, %w: %w", t.Synchronizer.Cached(), ErrNetIO, err)
	}

	return nil
}

// Recv 接收消息事件
func (t *Transceiver) Recv(ctx context.Context) (IEvent, error) {
	t.recvMutex.Lock()
	defer t.recvMutex.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	if t.Conn == nil {
		return IEvent{}, errors.New("gtp: setting Conn is nil")
	}

	if t.Decoder == nil {
		return IEvent{}, errors.New("gtp: setting Decoder is nil")
	}

	for {
		select {
		case <-ctx.Done():
			return IEvent{}, fmt.Errorf("gtp: %w", context.Canceled)
		default:
		}

		// 解码消息
		mp, err := t.Decoder.Decode(t.Synchronizer)
		if err == nil {
			return IEvent{
				Flags: mp.Head.Flags,
				Seq:   mp.Head.Seq,
				Ack:   mp.Head.Ack,
				Msg:   mp.Msg,
			}, t.Synchronizer.Ack(mp.Head.Ack)
		}

		if !errors.Is(err, codec.ErrDataNotEnough) {
			return IEvent{}, fmt.Errorf("gtp: decode msg-packet failed, %w", err)
		}

		// 设置链路超时时间
		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return IEvent{}, fmt.Errorf("gtp: set conn recv timeout failed, %w: %w", ErrNetIO, err)
			}
		}

		// 从链路读取消息
		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return IEvent{}, fmt.Errorf("gtp: recv msg-packet failed, %w, %w: %w", err, ErrNetIO, err)
		}
	}
}

// Renew 刷新链路
func (t *Transceiver) Renew(conn net.Conn, remoteRecvSeq uint32) (sendReq, recvReq uint32, err error) {
	if conn == nil {
		return 0, 0, fmt.Errorf("%w, conn is nil", ErrRenewConn)
	}

	// 同步对端时序
	if err = t.Synchronizer.Synchronization(remoteRecvSeq); err != nil {
		return 0, 0, fmt.Errorf("%w, synchronize sequence failed, %s", ErrRenewConn, err)
	}

	// 切换连接
	if t.Conn != nil {
		t.Conn.Close()
	}
	t.Conn = conn

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
}

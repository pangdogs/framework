package transport

import (
	"errors"
	"fmt"
	"io"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/codec"
	"net"
	"os"
	"sync"
	"time"
)

var (
	ErrRenewConn     = errors.New("renew conn")          // 刷新链路错误
	ErrUnexpectedSeq = errors.New("unexpected sequence") // 收到非预期的消息序号，表示序号不连续
	ErrDiscardSeq    = errors.New("discard sequence")    // 收到已过期的消息序号，表示次消息已收到过
	ErrNetIO         = errors.New("net i/o")             // 网络io错误
	ErrTimeout       = os.ErrDeadlineExceeded            // 网络io超时
	ErrClosed        = os.ErrClosed                      // 网络链路已关闭
	ErrUnexpectedEOF = io.ErrUnexpectedEOF               // 非预期的io结束
	ErrEOF           = io.EOF                            // io结束
)

// Transceiver 消息事件收发器，线程安全
type Transceiver struct {
	Conn                 net.Conn       // 网络连接
	Encoder              codec.IEncoder // 消息包编码器
	Decoder              codec.IDecoder // 消息包解码器
	Timeout              time.Duration  // 网络io超时时间
	Buffer               Buffer         // 缓存
	sendMutex, recvMutex sync.Mutex     // 发送与接收消息锁
}

// Send 发送消息
func (t *Transceiver) Send(me Event[gtp.MsgReader]) error {
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("setting Conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("setting Encoder is nil")
	}

	if t.Buffer == nil {
		return errors.New("setting Buffer is nil")
	}

	// 编码消息
	if err := t.Encoder.StuffTo(t.Buffer, me.Flags, me.Msg); err != nil {
		return fmt.Errorf("stuff msg failed, %w", err)
	}

	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("set conn send timeout failed, cached: %d, %w: %w", t.Buffer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Buffer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("send msg-packet failed, cached: %d, %w: %w", t.Buffer.Cached(), ErrNetIO, err)
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
	return t.Send(rstErr.Event().Pack())
}

// Resend 重新发送未完整发送的消息事件
func (t *Transceiver) Resend() error {
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("setting Conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("setting Encoder is nil")
	}

	if t.Buffer == nil {
		return errors.New("setting Buffer is nil")
	}

	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("set conn send timeout failed, cached: %d, %w: %w", t.Buffer.Cached(), ErrNetIO, err)
		}
	}

	// 数据写入链路
	if _, err := t.Buffer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("resend msg-packet failed, cached: %d, %w: %w", t.Buffer.Cached(), ErrNetIO, err)
	}

	return nil
}

// Recv 接收消息事件
func (t *Transceiver) Recv() (Event[gtp.Msg], error) {
	t.recvMutex.Lock()
	defer t.recvMutex.Unlock()

	if t.Conn == nil {
		return Event[gtp.Msg]{}, errors.New("setting Conn is nil")
	}

	if t.Decoder == nil {
		return Event[gtp.Msg]{}, errors.New("setting Decoder is nil")
	}

	for {
		// 解码消息
		mp, err := t.Decoder.Fetch(t.Buffer)
		if err == nil {
			return Event[gtp.Msg]{
				Flags: mp.Head.Flags,
				Seq:   mp.Head.Seq,
				Ack:   mp.Head.Ack,
				Msg:   mp.Msg,
			}, t.Buffer.Ack(mp.Head.Ack)
		}

		if !errors.Is(err, codec.ErrBufferNotEnough) {
			return Event[gtp.Msg]{}, fmt.Errorf("fetch msg-packet failed, %w", err)
		}

		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return Event[gtp.Msg]{}, fmt.Errorf("set conn recv timeout failed, %w: %w", ErrNetIO, err)
			}
		}

		// 从链路读取消息
		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return Event[gtp.Msg]{}, fmt.Errorf("recv msg-packet failed, %w, %w: %w", err, ErrNetIO, err)
		}
	}
}

// Renew 刷新链路
func (t *Transceiver) Renew(conn net.Conn, remoteRecvSeq uint32) (sendReq, recvReq uint32, err error) {
	if conn == nil {
		return 0, 0, fmt.Errorf("%w, conn is nil", ErrRenewConn)
	}

	// 同步对端时序
	if err = t.Buffer.Synchronization(remoteRecvSeq); err != nil {
		return 0, 0, fmt.Errorf("%w, synchronize sequence failed, %s", ErrRenewConn, err)
	}

	// 切换连接
	if t.Conn != nil {
		t.Conn.Close()
	}
	t.Conn = conn

	return t.Buffer.SendSeq(), t.Buffer.RecvSeq(), nil
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
	if t.Buffer != nil {
		t.Buffer.Clean()
	}
}

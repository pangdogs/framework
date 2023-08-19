package protocol

import (
	"errors"
	"fmt"
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"os"
	"sync"
	"time"
)

var (
	ErrUnexpectedSeq = errors.New("unexpected sequence") // 收到非预期的消息序号，表示序号不连续
	ErrDiscardSeq    = errors.New("discard sequence")    // 收到已过期的消息序号，表示次消息已收到过
	ErrNetIO         = errors.New("net i/o")             // 网络io类错误
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

// Send 发送消息事件
func (t *Transceiver) Send(e Event[transport.Msg]) error {
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	if t.Buffer == nil {
		return errors.New("buffer is nil")
	}

	// 编码消息
	if err := t.Encoder.StuffTo(t.Buffer, e.Flags, e.Msg); err != nil {
		return fmt.Errorf("stuff msg failed, %w", err)
	}

	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("set conn send timeout failed, %w", err)
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
	if ok := errors.As(err, &rstErr); !ok {
		rstErr = &RstError{Code: transport.Code_Reject}
		if err != nil {
			rstErr.Message = err.Error()
		}
	}
	return t.Send(PackEvent(RstErrToEvent(rstErr)))
}

// Resend 重新发送未完整发送的消息事件
func (t *Transceiver) Resend() error {
	t.sendMutex.Lock()
	defer t.sendMutex.Unlock()

	if t.Conn == nil {
		return errors.New("conn is nil")
	}

	if t.Encoder == nil {
		return errors.New("encoder is nil")
	}

	if t.Buffer == nil {
		return errors.New("buffer is nil")
	}

	if t.Timeout > 0 {
		if err := t.Conn.SetWriteDeadline(time.Now().Add(t.Timeout)); err != nil {
			return fmt.Errorf("set conn send timeout failed, %w", err)
		}
	}

	// 数据写入链路
	if _, err := t.Buffer.WriteTo(t.Conn); err != nil {
		return fmt.Errorf("resend msg-packet failed, cached: %d, %w: %w", t.Buffer.Cached(), ErrNetIO, err)
	}

	return nil
}

// Recv 接收消息事件
func (t *Transceiver) Recv() (Event[transport.Msg], error) {
	t.recvMutex.Lock()
	defer t.recvMutex.Unlock()

	if t.Conn == nil {
		return Event[transport.Msg]{}, errors.New("conn is nil")
	}

	if t.Decoder == nil {
		return Event[transport.Msg]{}, errors.New("decoder is nil")
	}

	for {
		// 解码消息
		mp, fetchErr := t.Decoder.Fetch(t.Buffer.Validation)
		if fetchErr == nil {
			return Event[transport.Msg]{
				Flags: mp.Head.Flags,
				Seq:   mp.Head.Seq,
				Ack:   mp.Head.Ack,
				Msg:   mp.Msg,
			}, t.Buffer.Ack(mp.Head.Ack)
		}

		if !errors.Is(fetchErr, codec.ErrBufferNotEnough) {
			return Event[transport.Msg]{}, fmt.Errorf("fetch msg-packet failed, %w", fetchErr)
		}

		if t.Timeout > 0 {
			if err := t.Conn.SetReadDeadline(time.Now().Add(t.Timeout)); err != nil {
				return Event[transport.Msg]{}, fmt.Errorf("set conn recv timeout failed, %w", err)
			}
		}

		// 从链路读取消息
		if _, err := t.Decoder.ReadFrom(t.Conn); err != nil {
			return Event[transport.Msg]{}, fmt.Errorf("recv msg-packet failed, %w, %w: %w", fetchErr, ErrNetIO, err)
		}
	}
}

// Renew 刷新链路
func (t *Transceiver) Renew(conn net.Conn, remoteRecvSeq uint32) (sendReq, recvReq uint32, err error) {
	if conn == nil {
		return 0, 0, errors.New("conn is nil")
	}

	// 同步对端时序
	if err = t.Buffer.Synchronization(remoteRecvSeq); err != nil {
		return 0, 0, fmt.Errorf("synchronize sequence failed, %s", err)
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

package transport

import (
	"errors"
	"kit.golaxy.org/plugins/gtp"
)

// Retry 网络io超时时重试
type Retry struct {
	Transceiver *Transceiver
	Times       int
}

// Send 重试发送
func (r Retry) Send(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrTimeout) {
		return err
	}
	for i := r.Times; i > 0; i-- {
		if err = r.Transceiver.Resend(); err != nil {
			if errors.Is(err, ErrTimeout) {
				continue
			}
		}
		break
	}
	return err
}

// Recv 重试接收
func (r Retry) Recv(e Event[gtp.Msg], err error) (Event[gtp.Msg], error) {
	if err == nil {
		return e, nil
	}
	if !errors.Is(err, ErrTimeout) && !errors.Is(err, ErrDiscardSeq) {
		return e, err
	}
	for i := r.Times; i > 0; {
		e, err = r.Transceiver.Recv()
		if err != nil {
			if errors.Is(err, ErrTimeout) {
				i--
				continue
			}
			if errors.Is(err, ErrDiscardSeq) {
				continue
			}
		}
		break
	}
	return e, err
}

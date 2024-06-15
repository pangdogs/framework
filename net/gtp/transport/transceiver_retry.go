package transport

import (
	"context"
	"errors"
	"fmt"
)

// Retry 网络io超时时重试
type Retry struct {
	Transceiver *Transceiver
	Times       int
	Ctx         context.Context
}

// Send 重试发送
func (r Retry) Send(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrTimeout) {
		return err
	}
	ctx := r.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	for i := r.Times; i > 0; i-- {
		select {
		case <-ctx.Done():
			return fmt.Errorf("gtp: %w", context.Canceled)
		default:
		}
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
func (r Retry) Recv(e IEvent, err error) (IEvent, error) {
	if err == nil {
		return e, nil
	}
	if !errors.Is(err, ErrTimeout) && !errors.Is(err, ErrDiscardSeq) {
		return e, err
	}
	for i := r.Times; i > 0; {
		e, err = r.Transceiver.Recv(r.Ctx)
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

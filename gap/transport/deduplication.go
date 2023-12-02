package transport

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrDiscardSeq = errors.New("discard sequence") // 收到已过期的消息序号，表示次消息已收到过
)

// MakeDeduplication 创建消息去重器
func MakeDeduplication() Deduplication {
	return Deduplication{
		selfSeq:      time.Now().UnixMicro(),
		remoteSeqMap: make(map[string]*int64),
	}
}

// Deduplication 消息去重器
type Deduplication struct {
	selfSeq      int64
	remoteSeqMap map[string]*int64
	remoteMutex  sync.Mutex
}

// MakeSeq 创建序号
func (d *Deduplication) MakeSeq() int64 {
	return atomic.AddInt64(&d.selfSeq, 1)
}

// ValidateSeq 验证序号
func (d *Deduplication) ValidateSeq(src string, seq int64) error {
	d.remoteMutex.Lock()
	defer d.remoteMutex.Unlock()

	remoteSeq, ok := d.remoteSeqMap[src]
	if !ok {
		d.remoteSeqMap[src] = &seq
		return nil
	}

	if seq <= *remoteSeq {
		return ErrDiscardSeq
	}

	*remoteSeq = seq
	return nil
}

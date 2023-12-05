package concurrent

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrDiscardSeq = errors.New("discard sequence") // 丢弃的序号
)

// IDeduplication 去重器接口
type IDeduplication interface {
	// MakeSeq 创建序号
	MakeSeq() int64
	// ValidateSeq 验证序号
	ValidateSeq(remote string, seq int64) error
	// Remove 删除对端
	Remove(remote string)
}

// MakeDeduplication 创建去重器
func MakeDeduplication() Deduplication {
	return Deduplication{
		localSeq:     time.Now().UnixMicro(),
		remoteSeqMap: make(map[string]*int64),
	}
}

// Deduplication 去重器，用于保持幂等性
type Deduplication struct {
	localSeq     int64
	remoteSeqMap map[string]*int64
	remoteMutex  sync.Mutex
}

// MakeSeq 创建序号
func (d *Deduplication) MakeSeq() int64 {
	return atomic.AddInt64(&d.localSeq, 1)
}

// ValidateSeq 验证序号
func (d *Deduplication) ValidateSeq(remote string, seq int64) error {
	d.remoteMutex.Lock()
	defer d.remoteMutex.Unlock()

	remoteSeq, ok := d.remoteSeqMap[remote]
	if !ok {
		d.remoteSeqMap[remote] = &seq
		return nil
	}

	if seq <= *remoteSeq {
		return ErrDiscardSeq
	}

	*remoteSeq = seq
	return nil
}

// Remove 删除对端
func (d *Deduplication) Remove(remote string) {
	d.remoteMutex.Lock()
	defer d.remoteMutex.Unlock()

	delete(d.remoteSeqMap, remote)
}

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
		localSeq: time.Now().UnixMicro(),
	}
}

type _RemoteSeq struct {
	sync.Mutex
	Seq int64
}

// Deduplication 去重器，用于保持幂等性
type Deduplication struct {
	localSeq       int64
	remoteSeqMap   map[string]*_RemoteSeq
	remoteSeqMutex sync.RWMutex
}

// MakeSeq 创建序号
func (d *Deduplication) MakeSeq() int64 {
	return atomic.AddInt64(&d.localSeq, 1)
}

// ValidateSeq 验证序号
func (d *Deduplication) ValidateSeq(remote string, seq int64) error {
	d.remoteSeqMutex.RLock()
	remoteSeq, ok := d.remoteSeqMap[remote]
	d.remoteSeqMutex.RUnlock()
	if !ok {
		d.remoteSeqMutex.Lock()
		remoteSeq, ok = d.remoteSeqMap[remote]
		if !ok {
			remoteSeq = &_RemoteSeq{
				Seq: seq,
			}
			d.remoteSeqMap[remote] = remoteSeq
		}
		d.remoteSeqMutex.Unlock()
	}

	remoteSeq.Lock()
	if seq <= remoteSeq.Seq {
		return ErrDiscardSeq
	}
	remoteSeq.Seq = seq
	remoteSeq.Unlock()

	return nil
}

// Remove 删除对端
func (d *Deduplication) Remove(remote string) {
	d.remoteSeqMutex.Lock()
	delete(d.remoteSeqMap, remote)
	d.remoteSeqMutex.Unlock()
}

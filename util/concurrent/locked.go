package concurrent

import (
	"kit.golaxy.org/golaxy/util/generic"
	"sync"
)

func MakeLocked[T any](obj T) Locked[T] {
	return Locked[T]{
		object: obj,
	}
}

type Locked[T any] struct {
	object T
	mutex  sync.RWMutex
}

func (l *Locked[T]) AutoLock(fun generic.Action1[*T]) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	fun.Exec(&l.object)
}

func (l *Locked[T]) AutoRLock(fun generic.Action1[*T]) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	fun.Exec(&l.object)
}

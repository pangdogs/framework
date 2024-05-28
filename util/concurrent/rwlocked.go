package concurrent

import (
	"git.golaxy.org/core/utils/generic"
	"sync"
)

func MakeRWLocked[T any](obj T) RWLocked[T] {
	return RWLocked[T]{
		object: obj,
	}
}

func NewRWLocked[T any](obj T) *RWLocked[T] {
	return &RWLocked[T]{
		object: obj,
	}
}

type RWLocked[T any] struct {
	object T
	mutex  sync.RWMutex
}

func (l *RWLocked[T]) AutoLock(fun generic.Action1[*T]) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	fun.Exec(&l.object)
}

func (l *RWLocked[T]) AutoRLock(fun generic.Action1[*T]) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	fun.Exec(&l.object)
}

package concurrent

import "git.golaxy.org/core/utils/generic"

type IMapEachElement[K comparable, V any] interface {
	Each(fun generic.Action2[K, V])
}

func MakeLockedMap[K comparable, V any](size int) LockedMap[K, V] {
	return LockedMap[K, V]{
		RWLocked: MakeRWLocked(make(map[K]V, size)),
	}
}

func NewLockedMap[K comparable, V any](size int) *LockedMap[K, V] {
	return &LockedMap[K, V]{
		RWLocked: MakeRWLocked(make(map[K]V, size)),
	}
}

type LockedMap[K comparable, V any] struct {
	RWLocked[map[K]V]
}

func (lm *LockedMap[K, V]) Insert(k K, v V) {
	lm.AutoLock(func(m *map[K]V) {
		(*m)[k] = v
	})
}

func (lm *LockedMap[K, V]) Delete(k K) {
	lm.AutoLock(func(m *map[K]V) {
		delete(*m, k)
	})
}

func (lm *LockedMap[K, V]) Get(k K) (v V, ok bool) {
	lm.AutoRLock(func(m *map[K]V) {
		v, ok = (*m)[k]
	})
	return
}

func (lm *LockedMap[K, V]) Len() (l int) {
	lm.AutoRLock(func(m *map[K]V) {
		l = len(*m)
	})
	return
}

func (lm *LockedMap[K, V]) Each(fun generic.Action2[K, V]) {
	lm.AutoRLock(func(m *map[K]V) {
		for k, v := range *m {
			fun.Exec(k, v)
		}
	})
}

func (lm *LockedMap[K, V]) Range(fun generic.Func2[K, V, bool]) {
	lm.AutoRLock(func(m *map[K]V) {
		for k, v := range *m {
			if !fun.Exec(k, v) {
				return
			}
		}
	})
}

package concurrent

import (
	"github.com/elliotchance/pie/v2"
)

func MakeLockedSlice[T any](len, cap int) LockedSlice[T] {
	return LockedSlice[T]{
		Locked: MakeLocked[[]T](make([]T, len, cap)),
	}
}

type LockedSlice[T any] struct {
	Locked[[]T]
}

func (ls *LockedSlice[T]) Insert(idx int, values ...T) {
	ls.AutoLock(func(s *[]T) {
		*s = pie.Insert(*s, idx, values...)
	})
}

func (ls *LockedSlice[T]) Append(values ...T) {
	ls.AutoLock(func(s *[]T) {
		*s = pie.Insert(ls.object, len(ls.object), values...)
	})
}

func (ls *LockedSlice[T]) Delete(idx ...int) {
	ls.AutoLock(func(s *[]T) {
		*s = pie.Delete(ls.object, idx...)
	})
}

func (ls *LockedSlice[T]) Len() (l int) {
	ls.AutoRLock(func(s *[]T) {
		l = len(*s)
	})
	return
}

package concurrent

import (
	"git.golaxy.org/core/utils/generic"
	"github.com/elliotchance/pie/v2"
)

func MakeLockedSlice[T any](len, cap int) LockedSlice[T] {
	return LockedSlice[T]{
		RWLocked: MakeRWLocked[[]T](make([]T, len, cap)),
	}
}

func NewLockedSlice[T any](len, cap int) *LockedSlice[T] {
	return &LockedSlice[T]{
		RWLocked: MakeRWLocked[[]T](make([]T, len, cap)),
	}
}

type LockedSlice[T any] struct {
	RWLocked[[]T]
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

func (ls *LockedSlice[T]) Range(fun generic.Func1[T, bool]) {
	ls.AutoRLock(func(s *[]T) {
		for i := range *s {
			if !fun.Exec((*s)[i]) {
				return
			}
		}
	})
}

func (ls *LockedSlice[T]) Each(fun generic.Action1[T]) {
	ls.AutoRLock(func(s *[]T) {
		for i := range *s {
			fun.Exec((*s)[i])
		}
	})
}

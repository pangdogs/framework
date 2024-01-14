package concurrent

import (
	"context"
	"git.golaxy.org/core/util/types"
	"time"
)

func newTask[T Resp](fs *Futures, resp T) _ITask {
	ctx, cancel := context.WithCancel(context.Background())

	task := &_Task[T]{
		future: Future{
			Finish:  ctx,
			Id:      fs.makeId(),
			futures: fs,
		},
		resp: resp,
		stop: cancel,
	}
	fs.tasks.Store(task.future.Id, task)

	return task
}

type _ITask interface {
	Future() Future
	Run(ctx context.Context, timeout time.Duration)
	Resolve(ret Ret[any]) error
}

type _Task[T Resp] struct {
	future Future
	resp   T
	stop   context.CancelFunc
}

func (t *_Task[T]) Future() Future {
	return t.future
}

func (t *_Task[T]) Run(ctx context.Context, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-t.future.futures.Ctx.Done():
		t.future.futures.Resolve(t.future.Id, Ret[any]{Error: ErrFuturesClosed})
	case <-ctx.Done():
		t.future.futures.Resolve(t.future.Id, Ret[any]{Error: ErrFutureCanceled})
	case <-timer.C:
		t.future.futures.Resolve(t.future.Id, Ret[any]{Error: ErrFutureTimeout})
	case <-t.future.Finish.Done():
		return
	}
}

func (t *_Task[T]) Resolve(ret Ret[any]) (retErr error) {
	t.stop()

	defer func() {
		if err := types.Panic2Err(recover()); err != nil {
			retErr = err
		}
	}()

	return t.resp.Push(ret)
}

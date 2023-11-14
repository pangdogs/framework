package concurrent

import (
	"golang.org/x/net/context"
	"time"
)

func newTask[T Resp](fs *Futures, resp T) _ITask {
	ctx, cancel := context.WithCancel(context.Background())

	task := &_Task[T]{
		future: Future{
			Ctx:     ctx,
			Id:      fs.MakeId(),
			futures: fs,
		},
		resp:   resp,
		cancel: cancel,
	}
	fs.tasks.Store(task.future.Id, task)

	return task
}

type _ITask interface {
	Future() Future
	Run(ctx context.Context, timeout time.Duration)
	Reply(ret Ret[any]) error
}

type _Task[T Resp] struct {
	future Future
	resp   T
	cancel context.CancelFunc
}

func (t *_Task[T]) Future() Future {
	return t.future
}

func (t *_Task[T]) Run(ctx context.Context, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-t.future.futures.Ctx.Done():
		t.future.futures.Dispatching(t.future.Id, Ret[any]{Error: ErrFuturesClosed})
	case <-ctx.Done():
		t.future.futures.Dispatching(t.future.Id, Ret[any]{Error: ErrFutureCancelled})
	case <-timer.C:
		t.future.futures.Dispatching(t.future.Id, Ret[any]{Error: ErrFutureTimeout})
	case <-t.future.Ctx.Done():
		return
	}
}

func (t *_Task[T]) Reply(ret Ret[any]) error {
	t.cancel()
	return t.resp.Push(ret)
}

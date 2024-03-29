package dserv

import (
	"context"
	"github.com/elliotchance/pie/v2"
)

func (d *_DistService) newMsgWatcher(ctx context.Context, handler RecvMsgHandler) *_MsgWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_MsgWatcher{
		Context:        ctx,
		terminate:      cancel,
		terminatedChan: make(chan struct{}),
		distributed:    d,
		handler:        handler,
	}
	d.msgWatchers.Append(watcher)

	d.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _MsgWatcher struct {
	context.Context
	terminate      context.CancelFunc
	terminatedChan chan struct{}
	distributed    *_DistService
	handler        RecvMsgHandler
}

func (w *_MsgWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminatedChan
}

func (w *_MsgWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.distributed.wg.Done()
		close(w.terminatedChan)
	}()

	select {
	case <-w.Done():
	case <-w.distributed.ctx.Done():
	}

	w.distributed.msgWatchers.AutoLock(func(watchers *[]*_MsgWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_MsgWatcher) bool {
			return other == w
		})
	})
}

package dist

import (
	"github.com/elliotchance/pie/v2"
	"golang.org/x/net/context"
)

func (d *_Distributed) newMsgWatcher(ctx context.Context, handler RecvMsgHandler) *_MsgWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_MsgWatcher{
		Context:     ctx,
		cancel:      cancel,
		stoppedChan: make(chan struct{}),
		distributed: d,
		handler:     handler,
	}
	d.msgWatchers.Append(watcher)

	d.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _MsgWatcher struct {
	context.Context
	cancel      context.CancelFunc
	stoppedChan chan struct{}
	distributed *_Distributed
	handler     RecvMsgHandler
}

func (w *_MsgWatcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_MsgWatcher) mainLoop() {
	defer func() {
		w.cancel()
		w.distributed.wg.Done()
		close(w.stoppedChan)
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

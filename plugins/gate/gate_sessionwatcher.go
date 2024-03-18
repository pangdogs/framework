package gate

import (
	"context"
	"github.com/elliotchance/pie/v2"
)

func (g *_Gate) newSessionWatcher(ctx context.Context, handler SessionStateChangedHandler) *_SessionWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_SessionWatcher{
		Context:     ctx,
		terminate:   cancel,
		stoppedChan: make(chan struct{}),
		gate:        g,
		handler:     handler,
	}
	g.sessionWatchers.Append(watcher)

	g.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _SessionWatcher struct {
	context.Context
	terminate   context.CancelFunc
	stoppedChan chan struct{}
	gate        *_Gate
	handler     SessionStateChangedHandler
}

func (w *_SessionWatcher) Stop() <-chan struct{} {
	w.terminate()
	return w.stoppedChan
}

func (w *_SessionWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.gate.wg.Done()
		close(w.stoppedChan)
	}()

	select {
	case <-w.Done():
	case <-w.gate.ctx.Done():
	}

	w.gate.sessionWatchers.AutoLock(func(watchers *[]*_SessionWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_SessionWatcher) bool {
			return other == w
		})
	})
}

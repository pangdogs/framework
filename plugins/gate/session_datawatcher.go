package gate

import (
	"context"
	"github.com/elliotchance/pie/v2"
)

func (s *_Session) newDataWatcher(ctx context.Context, handler SessionRecvDataHandler) *_DataWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_DataWatcher{
		Context:        ctx,
		terminate:      cancel,
		terminatedChan: make(chan struct{}),
		session:        s,
		handler:        handler,
	}
	s.dataWatchers.Append(watcher)

	s.gate.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _DataWatcher struct {
	context.Context
	terminate      context.CancelFunc
	terminatedChan chan struct{}
	session        *_Session
	handler        SessionRecvDataHandler
}

func (w *_DataWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminatedChan
}

func (w *_DataWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.session.gate.wg.Done()
		close(w.terminatedChan)
	}()

	select {
	case <-w.Done():
	case <-w.session.Done():
	}

	w.session.dataWatchers.AutoLock(func(watchers *[]*_DataWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_DataWatcher) bool {
			return other == w
		})
	})
}

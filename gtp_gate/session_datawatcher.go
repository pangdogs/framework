package gtp_gate

import (
	"github.com/elliotchance/pie/v2"
	"golang.org/x/net/context"
)

func (s *_Session) newEventWatcher(ctx context.Context, handler RecvEventHandler) *_EventWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_EventWatcher{
		Context:     ctx,
		session:     s,
		cancel:      cancel,
		stoppedChan: make(chan struct{}),
		handler:     handler,
	}
	s.eventWatchers.Append(watcher)

	s.gate.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _EventWatcher struct {
	context.Context
	cancel      context.CancelFunc
	stoppedChan chan struct{}
	session     *_Session
	handler     RecvEventHandler
}

func (w *_EventWatcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_EventWatcher) mainLoop() {
	defer func() {
		w.cancel()
		w.session.gate.wg.Done()
		close(w.stoppedChan)
	}()

	select {
	case <-w.Done():
	case <-w.session.Done():
	}

	w.session.eventWatchers.AutoLock(func(watchers *[]*_EventWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_EventWatcher) bool {
			return other == w
		})
	})
}

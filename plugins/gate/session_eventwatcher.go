package gate

import (
	"context"
	"github.com/elliotchance/pie/v2"
)

func (s *_Session) newEventWatcher(ctx context.Context, handler SessionRecvEventHandler) *_EventWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_EventWatcher{
		Context:        ctx,
		terminate:      cancel,
		terminatedChan: make(chan struct{}),
		session:        s,
		handler:        handler,
	}
	s.eventWatchers.Append(watcher)

	s.gate.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _EventWatcher struct {
	context.Context
	terminate      context.CancelFunc
	terminatedChan chan struct{}
	session        *_Session
	handler        SessionRecvEventHandler
}

func (w *_EventWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminatedChan
}

func (w *_EventWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.session.gate.wg.Done()
		close(w.terminatedChan)
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

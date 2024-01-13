package gtp_cli

import (
	"github.com/elliotchance/pie/v2"
	"golang.org/x/net/context"
)

func (c *Client) newEventWatcher(ctx context.Context, handler RecvEventHandler) *_EventWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_EventWatcher{
		Context:     ctx,
		cancel:      cancel,
		stoppedChan: make(chan struct{}),
		client:      c,
		handler:     handler,
	}
	c.eventWatchers.Append(watcher)

	c.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _EventWatcher struct {
	context.Context
	cancel      context.CancelFunc
	stoppedChan chan struct{}
	client      *Client
	handler     RecvEventHandler
}

func (w *_EventWatcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_EventWatcher) mainLoop() {
	defer func() {
		w.cancel()
		w.client.wg.Done()
		close(w.stoppedChan)
	}()

	select {
	case <-w.Done():
	case <-w.client.Done():
	}

	w.client.eventWatchers.AutoLock(func(watchers *[]*_EventWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_EventWatcher) bool {
			return other == w
		})
	})
}

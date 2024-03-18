package cli

import (
	"context"
	"github.com/elliotchance/pie/v2"
)

func (c *Client) newDataWatcher(ctx context.Context, handler RecvDataHandler) *_DataWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_DataWatcher{
		Context:        ctx,
		terminate:      cancel,
		terminatedChan: make(chan struct{}),
		client:         c,
		handler:        handler,
	}
	c.dataWatchers.Append(watcher)

	c.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _DataWatcher struct {
	context.Context
	terminate      context.CancelFunc
	terminatedChan chan struct{}
	client         *Client
	handler        RecvDataHandler
}

func (w *_DataWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminatedChan
}

func (w *_DataWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.client.wg.Done()
		close(w.terminatedChan)
	}()

	select {
	case <-w.Done():
	case <-w.client.Done():
	}

	w.client.dataWatchers.AutoLock(func(watchers *[]*_DataWatcher) {
		*watchers = pie.DropWhile(*watchers, func(other *_DataWatcher) bool {
			return other == w
		})
	})
}

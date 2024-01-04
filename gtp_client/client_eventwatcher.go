package gtp_client

import (
	"github.com/elliotchance/pie/v2"
	"golang.org/x/net/context"
)

func (c *Client) newDataWatcher(ctx context.Context, handler RecvDataHandler) *_DataWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_DataWatcher{
		Context:     ctx,
		client:      c,
		cancel:      cancel,
		stoppedChan: make(chan struct{}),
		handler:     handler,
	}
	c.dataWatchers.Append(watcher)

	c.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _DataWatcher struct {
	context.Context
	client      *Client
	cancel      context.CancelFunc
	stoppedChan chan struct{}
	handler     RecvDataHandler
}

func (w *_DataWatcher) Stop() <-chan struct{} {
	w.cancel()
	return w.stoppedChan
}

func (w *_DataWatcher) mainLoop() {
	defer func() {
		w.cancel()
		w.client.wg.Done()
		close(w.stoppedChan)
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

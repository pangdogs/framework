/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package cli

import (
	"context"
)

func (c *Client) newDataWatcher(ctx context.Context, handler RecvDataHandler) *_DataWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_DataWatcher{
		Context:    ctx,
		terminate:  cancel,
		terminated: make(chan struct{}),
		client:     c,
		handler:    handler,
	}
	c.dataWatchers.Append(watcher)

	c.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _DataWatcher struct {
	context.Context
	terminate  context.CancelFunc
	terminated chan struct{}
	client     *Client
	handler    RecvDataHandler
}

func (w *_DataWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminated
}

func (w *_DataWatcher) Terminated() <-chan struct{} {
	return w.terminated
}

func (w *_DataWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.client.wg.Done()
		close(w.terminated)
	}()

	select {
	case <-w.Done():
	case <-w.client.Done():
	}

	w.client.dataWatchers.Delete(func(exists *_DataWatcher) bool {
		return exists == w
	})
}

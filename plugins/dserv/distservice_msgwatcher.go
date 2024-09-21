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

package dserv

import (
	"context"
	"slices"
)

func (d *_DistService) newMsgWatcher(ctx context.Context, handler RecvMsgHandler) *_MsgWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_MsgWatcher{
		Context:     ctx,
		terminate:   cancel,
		terminated:  make(chan struct{}),
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
	terminate   context.CancelFunc
	terminated  chan struct{}
	distributed *_DistService
	handler     RecvMsgHandler
}

func (w *_MsgWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminated
}

func (w *_MsgWatcher) Terminated() <-chan struct{} {
	return w.terminated
}

func (w *_MsgWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.distributed.wg.Done()
		close(w.terminated)
	}()

	select {
	case <-w.Done():
	case <-w.distributed.ctx.Done():
	}

	w.distributed.msgWatchers.AutoLock(func(watchers *[]*_MsgWatcher) {
		*watchers = slices.DeleteFunc(*watchers, func(other *_MsgWatcher) bool {
			return other == w
		})
	})
}

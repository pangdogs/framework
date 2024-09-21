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

package gate

import (
	"context"
)

func (g *_Gate) newSessionWatcher(ctx context.Context, handler SessionStateChangedHandler) *_SessionWatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	watcher := &_SessionWatcher{
		Context:    ctx,
		terminate:  cancel,
		terminated: make(chan struct{}),
		gate:       g,
		handler:    handler,
	}
	g.sessionWatchers.Append(watcher)

	g.wg.Add(1)
	go watcher.mainLoop()

	return watcher
}

type _SessionWatcher struct {
	context.Context
	terminate  context.CancelFunc
	terminated chan struct{}
	gate       *_Gate
	handler    SessionStateChangedHandler
}

func (w *_SessionWatcher) Terminate() <-chan struct{} {
	w.terminate()
	return w.terminated
}

func (w *_SessionWatcher) Terminated() <-chan struct{} {
	return w.terminated
}

func (w *_SessionWatcher) mainLoop() {
	defer func() {
		w.terminate()
		w.gate.wg.Done()
		close(w.terminated)
	}()

	select {
	case <-w.Done():
	case <-w.gate.ctx.Done():
	}

	w.gate.sessionWatchers.DeleteOnce(func(exists *_SessionWatcher) bool {
		return exists == w
	})
}

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
	"errors"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/log"
	"go.uber.org/zap"
)

type (
	SessionEstablishedHandler = generic.DelegateVoid1[ISession] // 会话建立完成处理器
)

func (g *_Gate) addSessionWatcher(ctx context.Context, handler SessionEstablishedHandler) (async.Future, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-g.ctx.Done():
		return errors.New("gate: gate is terminating")
	default:
	}

	if !g.barrier.Join(1) {
		return errors.New("gate: gate is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-g.ctx.Done():
		}
		cancel()
	}()

	watcher := g.sessionWatcher.Add(handler, g.options.SessionWatcherInboxSize)

	go func() {
		defer g.barrier.Done()

		for {
			select {
			case <-ctx.Done():
				g.sessionWatcher.Delete(watcher)
				log.L(g.svcCtx).Debug("delete a session established watcher")
				return
			case session := <-watcher.Inbox:
				watcher.Handler.Call(g.svcCtx.AutoRecover(), g.svcCtx.ReportError(), func(panicError error) bool {
					if panicError != nil {
						addr := session.NetAddr()
						log.L(g.svcCtx).Error("handle session established panicked",
							zap.String("session_id", session.Id().String()),
							zap.String("user_id", session.UserId()),
							zap.String("token", session.Token()),
							zap.Int64("migrations", session.Migrations()),
							zap.String("local", addr.Local.String()),
							zap.String("remote", addr.Remote.String()),
							zap.Error(panicError))
					}
					return false
				}, session)
			}
		}
	}()

	log.L(g.svcCtx).Debug("add a session established watcher")
	return nil
}

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

package router

import (
	"errors"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"go.uber.org/zap"
)

// IDataIO 数据IO
type IDataIO interface {
	Send(data []byte) error
}

// IEventIO 事件IO
type IEventIO interface {
	Send(event transport.IEvent) error
}

type _GroupIO struct {
	group     *_Group
	barrier   generic.Barrier
	dataChan  *generic.UnboundedChannel[binaryutil.Bytes]
	eventChan *generic.UnboundedChannel[transport.IEvent]
}

func (io *_GroupIO) init(group *_Group) {
	io.group = group
	io.dataChan = generic.NewUnboundedChannel[binaryutil.Bytes]()
	io.eventChan = generic.NewUnboundedChannel[transport.IEvent]()
}

func (io *_GroupIO) sendLoop() {
loop:
	for {
		select {
		case <-io.group.expired:
			break loop

		case buff := <-io.dataChan.Out():
			if err := io.group.sendData(buff.Payload()); err != nil {
				log.L(io.group.router.svcCtx).Error("group send data failed",
					zap.String("group_name", io.group.Name()),
					zap.String("group_addr", io.group.ClientAddr()),
					zap.Error(err))
			}
			buff.Release()

		case event := <-io.eventChan.Out():
			if err := io.group.sendEvent(event); err != nil {
				log.L(io.group.router.svcCtx).Error("group send event failed",
					zap.String("group_name", io.group.Name()),
					zap.String("group_addr", io.group.ClientAddr()),
					zap.Error(err))
			}
		}
	}

	io.barrier.Close()
	io.barrier.Wait()

	io.dataChan.Close()
	io.eventChan.Close()

	for buff := range io.dataChan.Out() {
		if err := io.group.sendData(buff.Payload()); err != nil {
			log.L(io.group.router.svcCtx).Error("group send data failed",
				zap.String("group_name", io.group.Name()),
				zap.String("group_addr", io.group.ClientAddr()),
				zap.Error(err))
		}
		buff.Release()
	}

	for event := range io.eventChan.Out() {
		if err := io.group.sendEvent(event); err != nil {
			log.L(io.group.router.svcCtx).Error("group send event failed",
				zap.String("group_name", io.group.Name()),
				zap.String("group_addr", io.group.ClientAddr()),
				zap.Error(err))
		}
	}
}

type _GroupDataIO _GroupIO

func (io *_GroupDataIO) Send(data []byte) (err error) {
	if !io.barrier.Join(1) {
		return errors.New("router: group data i/o is terminating")
	}
	defer io.barrier.Done()

	io.dataChan.In() <- binaryutil.CloneBytes(true, data)
	return nil
}

type _GroupEventIO _GroupIO

func (io *_GroupEventIO) Send(event transport.IEvent) (err error) {
	if !io.barrier.Join(1) {
		return errors.New("router: group event i/o is terminating")
	}
	defer io.barrier.Done()

	io.eventChan.In() <- event
	return nil
}

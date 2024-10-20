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

package gtp

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/utils/concurrent"
	"reflect"
)

var (
	ErrNotDeclared = fmt.Errorf("%w: msg not declared", ErrGTP) // 消息未注册
)

// IMsgCreator 消息对象构建器接口
type IMsgCreator interface {
	// Declare 注册消息
	Declare(msg Msg)
	// Undeclare 取消注册消息
	Undeclare(msgId MsgId)
	// New 创建消息指针
	New(msgId MsgId) (Msg, error)
}

var msgCreator = NewMsgCreator()

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return msgCreator
}

func init() {
	DefaultMsgCreator().Declare(&MsgHello{})
	DefaultMsgCreator().Declare(&MsgECDHESecretKeyExchange{})
	DefaultMsgCreator().Declare(&MsgChangeCipherSpec{})
	DefaultMsgCreator().Declare(&MsgAuth{})
	DefaultMsgCreator().Declare(&MsgContinue{})
	DefaultMsgCreator().Declare(&MsgFinished{})
	DefaultMsgCreator().Declare(&MsgRst{})
	DefaultMsgCreator().Declare(&MsgHeartbeat{})
	DefaultMsgCreator().Declare(&MsgSyncTime{})
	DefaultMsgCreator().Declare(&MsgPayload{})
}

// NewMsgCreator 创建消息对象构建器
func NewMsgCreator() IMsgCreator {
	return &_MsgCreator{
		msgTypeMap: concurrent.MakeLockedMap[MsgId, reflect.Type](0),
	}
}

// _MsgCreator 消息对象构建器
type _MsgCreator struct {
	msgTypeMap concurrent.LockedMap[MsgId, reflect.Type]
}

// Declare 注册消息
func (c *_MsgCreator) Declare(msg Msg) {
	if msg == nil {
		exception.Panicf("%w: %w: msg is nil", ErrGTP, core.ErrArgs)
	}

	c.msgTypeMap.AutoLock(func(m *map[MsgId]reflect.Type) {
		if rtype, ok := (*m)[msg.MsgId()]; ok {
			exception.Panicf("%w: msg(%d) has already been declared by %q", ErrGTP, msg.MsgId(), types.FullNameRT(rtype))
		}
		(*m)[msg.MsgId()] = reflect.TypeOf(msg).Elem()
	})
}

// Undeclare 取消注册消息
func (c *_MsgCreator) Undeclare(msgId MsgId) {
	c.msgTypeMap.Delete(msgId)
}

// New 创建消息指针
func (c *_MsgCreator) New(msgId MsgId) (Msg, error) {
	rtype, ok := c.msgTypeMap.Get(msgId)
	if !ok {
		return nil, ErrNotDeclared
	}
	return reflect.New(rtype).Interface().(Msg), nil
}

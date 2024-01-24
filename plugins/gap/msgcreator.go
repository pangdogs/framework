package gap

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/plugins/util/concurrent"
	"reflect"
)

var (
	ErrNotRegistered = errors.New("gap: msg not registered") // 消息未注册
)

// IMsgCreator 消息对象构建器接口
type IMsgCreator interface {
	// Register 注册消息
	Register(msg Msg)
	// Deregister 取消注册消息
	Deregister(msgId MsgId)
	// New 创建消息指针
	New(msgId MsgId) (Msg, error)
}

var msgCreator = NewMsgCreator()

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return msgCreator
}

func init() {
	DefaultMsgCreator().Register(&MsgRPCRequest{})
	DefaultMsgCreator().Register(&MsgRPCReply{})
	DefaultMsgCreator().Register(&MsgOneWayRPC{})
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

// Register 注册消息
func (c *_MsgCreator) Register(msg Msg) {
	if msg == nil {
		panic(fmt.Errorf("%w: msg is nil", core.ErrArgs))
	}

	c.msgTypeMap.Insert(msg.MsgId(), reflect.TypeOf(msg).Elem())
}

// Deregister 取消注册消息
func (c *_MsgCreator) Deregister(msgId MsgId) {
	c.msgTypeMap.Delete(msgId)
}

// New 创建消息指针
func (c *_MsgCreator) New(msgId MsgId) (Msg, error) {
	rtype, ok := c.msgTypeMap.Get(msgId)
	if !ok {
		return nil, ErrNotRegistered
	}

	return reflect.New(rtype).Interface().(Msg), nil
}

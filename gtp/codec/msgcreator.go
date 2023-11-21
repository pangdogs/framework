package codec

import (
	"errors"
	"kit.golaxy.org/plugins/gtp"
	"reflect"
	"sync"
)

var (
	ErrMsgNotRegistered = errors.New("msg not registered") // 消息未注册
)

// IMsgCreator 消息对象构建器接口
type IMsgCreator interface {
	// Register 注册消息
	Register(msg gtp.Msg)
	// Deregister 取消注册消息
	Deregister(msgId gtp.MsgId)
	// Spawn 构建消息
	Spawn(msgId gtp.MsgId) (gtp.Msg, error)
}

var msgCreator = NewMsgCreator()

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return msgCreator
}

func init() {
	DefaultMsgCreator().Register(&gtp.MsgHello{})
	DefaultMsgCreator().Register(&gtp.MsgECDHESecretKeyExchange{})
	DefaultMsgCreator().Register(&gtp.MsgChangeCipherSpec{})
	DefaultMsgCreator().Register(&gtp.MsgAuth{})
	DefaultMsgCreator().Register(&gtp.MsgContinue{})
	DefaultMsgCreator().Register(&gtp.MsgFinished{})
	DefaultMsgCreator().Register(&gtp.MsgRst{})
	DefaultMsgCreator().Register(&gtp.MsgHeartbeat{})
	DefaultMsgCreator().Register(&gtp.MsgSyncTime{})
	DefaultMsgCreator().Register(&gtp.MsgPayload{})
}

// NewMsgCreator 创建消息对象构建器
func NewMsgCreator() IMsgCreator {
	return &_MsgCreator{
		msgTypeMap: make(map[gtp.MsgId]reflect.Type),
	}
}

// _MsgCreator 消息对象构建器
type _MsgCreator struct {
	sync.RWMutex
	msgTypeMap map[gtp.MsgId]reflect.Type
}

// Register 注册消息
func (c *_MsgCreator) Register(msg gtp.Msg) {
	c.Lock()
	defer c.Unlock()

	c.msgTypeMap[msg.MsgId()] = reflect.TypeOf(msg).Elem()
}

// Deregister 取消注册消息
func (c *_MsgCreator) Deregister(msgId gtp.MsgId) {
	c.Lock()
	defer c.Unlock()

	delete(c.msgTypeMap, msgId)
}

// Spawn 构建消息
func (c *_MsgCreator) Spawn(msgId gtp.MsgId) (gtp.Msg, error) {
	c.RLock()
	defer c.RUnlock()

	rtype, ok := c.msgTypeMap[msgId]
	if !ok {
		return nil, ErrMsgNotRegistered
	}

	return reflect.New(rtype).Interface().(gtp.Msg), nil
}

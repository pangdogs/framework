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

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return &msgCreator
}

var msgCreator = _MsgCreator{}

func init() {
	msgCreator.Register(&gtp.MsgHello{})
	msgCreator.Register(&gtp.MsgECDHESecretKeyExchange{})
	msgCreator.Register(&gtp.MsgChangeCipherSpec{})
	msgCreator.Register(&gtp.MsgAuth{})
	msgCreator.Register(&gtp.MsgContinue{})
	msgCreator.Register(&gtp.MsgFinished{})
	msgCreator.Register(&gtp.MsgRst{})
	msgCreator.Register(&gtp.MsgHeartbeat{})
	msgCreator.Register(&gtp.MsgSyncTime{})
	msgCreator.Register(&gtp.MsgPayload{})
}

// _MsgCreator 消息对象构建器
type _MsgCreator struct {
	msgTypeMap map[gtp.MsgId]reflect.Type
	mutex      sync.RWMutex
}

// Register 注册消息
func (c *_MsgCreator) Register(msg gtp.Msg) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.msgTypeMap == nil {
		c.msgTypeMap = map[gtp.MsgId]reflect.Type{}
	}

	c.msgTypeMap[msg.MsgId()] = reflect.TypeOf(msg).Elem()
}

// Deregister 取消注册消息
func (c *_MsgCreator) Deregister(msgId gtp.MsgId) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.msgTypeMap == nil {
		return
	}

	delete(c.msgTypeMap, msgId)
}

// Spawn 构建消息
func (c *_MsgCreator) Spawn(msgId gtp.MsgId) (gtp.Msg, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.msgTypeMap == nil {
		return nil, ErrMsgNotRegistered
	}

	rtype, ok := c.msgTypeMap[msgId]
	if !ok {
		return nil, ErrMsgNotRegistered
	}

	return reflect.New(rtype).Interface().(gtp.Msg), nil
}

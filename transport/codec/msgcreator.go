package codec

import (
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"reflect"
)

// IMsgCreator 消息对象构建器接口
type IMsgCreator interface {
	// Spawn 构建消息
	Spawn(msgId transport.MsgId) (transport.Msg, error)
}

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return &msgCreator
}

var msgCreator = MsgCreator{}

func init() {
	msgCreator.Register(&transport.MsgHello{})
	msgCreator.Register(&transport.MsgECDHESecretKeyExchange{})
	msgCreator.Register(&transport.MsgChangeCipherSpec{})
	msgCreator.Register(&transport.MsgAuth{})
	msgCreator.Register(&transport.MsgContinue{})
	msgCreator.Register(&transport.MsgFinished{})
	msgCreator.Register(&transport.MsgRst{})
	msgCreator.Register(&transport.MsgHeartbeat{})
	msgCreator.Register(&transport.MsgSyncTime{})
	msgCreator.Register(&transport.MsgPayload{})
}

// MsgCreator 消息对象构建器
type MsgCreator struct {
	msgTypeMap map[transport.MsgId]reflect.Type
}

// Register 注册消息
func (c *MsgCreator) Register(msg transport.Msg) {
	if c.msgTypeMap == nil {
		c.msgTypeMap = map[transport.MsgId]reflect.Type{}
	}
	c.msgTypeMap[msg.MsgId()] = reflect.TypeOf(msg).Elem()
}

// Deregister 取消注册消息
func (c *MsgCreator) Deregister(msgId transport.MsgId) {
	if c.msgTypeMap == nil {
		return
	}
	delete(c.msgTypeMap, msgId)
}

// Spawn 构建消息
func (c *MsgCreator) Spawn(msgId transport.MsgId) (transport.Msg, error) {
	if c.msgTypeMap == nil {
		return nil, fmt.Errorf("msg %d not registered", msgId)
	}
	rtype, ok := c.msgTypeMap[msgId]
	if !ok {
		return nil, fmt.Errorf("msg %d not registered", msgId)
	}
	return reflect.New(rtype).Interface().(transport.Msg), nil
}

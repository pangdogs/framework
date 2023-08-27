package codec

import (
	"fmt"
	"kit.golaxy.org/plugins/gtp"
	"reflect"
)

// IMsgCreator 消息对象构建器接口
type IMsgCreator interface {
	// Spawn 构建消息
	Spawn(msgId gtp.MsgId) (gtp.Msg, error)
}

// DefaultMsgCreator 默认消息对象构建器
func DefaultMsgCreator() IMsgCreator {
	return &msgCreator
}

var msgCreator = MsgCreator{}

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

// MsgCreator 消息对象构建器
type MsgCreator struct {
	msgTypeMap map[gtp.MsgId]reflect.Type
}

// Register 注册消息
func (c *MsgCreator) Register(msg gtp.Msg) {
	if c.msgTypeMap == nil {
		c.msgTypeMap = map[gtp.MsgId]reflect.Type{}
	}
	c.msgTypeMap[msg.MsgId()] = reflect.TypeOf(msg).Elem()
}

// Deregister 取消注册消息
func (c *MsgCreator) Deregister(msgId gtp.MsgId) {
	if c.msgTypeMap == nil {
		return
	}
	delete(c.msgTypeMap, msgId)
}

// Spawn 构建消息
func (c *MsgCreator) Spawn(msgId gtp.MsgId) (gtp.Msg, error) {
	if c.msgTypeMap == nil {
		return nil, fmt.Errorf("msg %d not registered", msgId)
	}
	rtype, ok := c.msgTypeMap[msgId]
	if !ok {
		return nil, fmt.Errorf("msg %d not registered", msgId)
	}
	return reflect.New(rtype).Interface().(gtp.Msg), nil
}

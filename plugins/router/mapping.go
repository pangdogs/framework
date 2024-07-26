package router

import (
	"context"
	"git.golaxy.org/core/ec"
	"git.golaxy.org/framework/plugins/gate"
)

// IMapping 映射
type IMapping interface {
	context.Context
	// GetEntity 获取实体
	GetEntity() ec.ConcurrentEntity
	// GetSession 获取会话
	GetSession() gate.ISession
	// GetCliAddr 获取客户端地址
	GetCliAddr() string
}

type _Mapping struct {
	context.Context
	terminate context.CancelFunc
	router    *_Router
	entity    ec.ConcurrentEntity
	session   gate.ISession
	cliAddr   string
}

// GetEntity 获取实体
func (m *_Mapping) GetEntity() ec.ConcurrentEntity {
	return m.entity
}

// GetSession 获取会话
func (m *_Mapping) GetSession() gate.ISession {
	return m.session
}

// GetCliAddr 获取客户端地址
func (m *_Mapping) GetCliAddr() string {
	return m.cliAddr
}

func (m *_Mapping) mainLoop() {
	select {
	case <-m.Done():
		return
	case <-m.session.Done():
		m.router.CleanSession(m.session.GetId())
	}
}

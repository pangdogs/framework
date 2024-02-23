package gate

import (
	"errors"
	"git.golaxy.org/core/util/option"
)

// SessionSettings 会话设置
type SessionSettings struct {
	session  *_Session
	settings []option.Setting[_SessionOptions]
}

// StateChangedHandler 会话状态变化的处理器
func (s SessionSettings) StateChangedHandler(handler SessionStateChangedHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.StateChangedHandler(handler))
	return s
}

// SendDataChanSize 发送数据的channel的大小，<=0表示不使用channel
func (s SessionSettings) SendDataChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.SendDataChanSize(size))
	return s
}

// RecvDataChanSize 接收数据的channel的大小，<=0表示不使用channel
func (s SessionSettings) RecvDataChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvDataChanSize(size))
	return s
}

// SendEventChanSize 发送自定义事件的channel的大小，<=0表示不使用channel
func (s SessionSettings) SendEventChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.SendEventChanSize(size))
	return s
}

// RecvEventChanSize 接收自定义事件的channel的大小，<=0表示不使用channel
func (s SessionSettings) RecvEventChanSize(size int) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvEventChanSize(size))
	return s
}

// RecvDataHandler 接收的数据的处理器
func (s SessionSettings) RecvDataHandler(handler SessionRecvDataHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvDataHandler(handler))
	return s
}

// RecvEventHandler 接收自定义事件的处理器
func (s SessionSettings) RecvEventHandler(handler SessionRecvEventHandler) SessionSettings {
	s.settings = append(s.settings, sessionWith.RecvEventHandler(handler))
	return s
}

// Change 执行修改
func (s SessionSettings) Change() error {
	if s.session == nil {
		panic(errors.New("setting session is nil"))
	}

	s.session.Lock()
	defer s.session.Unlock()

	switch s.session.state {
	case SessionState_Birth, SessionState_Handshake, SessionState_Confirmed:
		break
	default:
		return errors.New("incorrect session state")
	}

	option.Change(&s.session.options, s.settings...)

	return nil
}

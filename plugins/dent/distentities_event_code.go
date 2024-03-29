// Code generated by eventcode --decl_file=distentities_event.go gen_event --package=dent --default_export=false; DO NOT EDIT.

package dent

import (
	"fmt"
	event "git.golaxy.org/core/event"
	iface "git.golaxy.org/core/util/iface"
	"git.golaxy.org/core/ec"
)

type iAutoEventDistEntityOnline interface {
	EventDistEntityOnline() event.IEvent
}

func BindEventDistEntityOnline(auto iAutoEventDistEntityOnline, subscriber EventDistEntityOnline, priority ...int32) event.Hook {
	if auto == nil {
		panic(fmt.Errorf("%w: %w: auto is nil", event.ErrEvent, event.ErrArgs))
	}
	return event.BindEvent[EventDistEntityOnline](auto.EventDistEntityOnline(), subscriber, priority...)
}

func emitEventDistEntityOnline(auto iAutoEventDistEntityOnline, entity ec.Entity) {
	if auto == nil {
		panic(fmt.Errorf("%w: %w: auto is nil", event.ErrEvent, event.ErrArgs))
	}
	event.UnsafeEvent(auto.EventDistEntityOnline()).Emit(func(subscriber iface.Cache) bool {
		iface.Cache2Iface[EventDistEntityOnline](subscriber).OnDistEntityOnline(entity)
		return true
	})
}

type iAutoEventDistEntityOffline interface {
	EventDistEntityOffline() event.IEvent
}

func BindEventDistEntityOffline(auto iAutoEventDistEntityOffline, subscriber EventDistEntityOffline, priority ...int32) event.Hook {
	if auto == nil {
		panic(fmt.Errorf("%w: %w: auto is nil", event.ErrEvent, event.ErrArgs))
	}
	return event.BindEvent[EventDistEntityOffline](auto.EventDistEntityOffline(), subscriber, priority...)
}

func emitEventDistEntityOffline(auto iAutoEventDistEntityOffline, entity ec.Entity) {
	if auto == nil {
		panic(fmt.Errorf("%w: %w: auto is nil", event.ErrEvent, event.ErrArgs))
	}
	event.UnsafeEvent(auto.EventDistEntityOffline()).Emit(func(subscriber iface.Cache) bool {
		iface.Cache2Iface[EventDistEntityOffline](subscriber).OnDistEntityOffline(entity)
		return true
	})
}

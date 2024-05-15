//go:generate go run git.golaxy.org/core/event/eventcode gen_event --default_export=false
//go:generate go run git.golaxy.org/core/event/eventcode gen_eventtab --name=distEntitiesEventTab
package dent

import (
	"git.golaxy.org/core/ec"
)

// EventDistEntityOnline 事件：分布式实体上线
type EventDistEntityOnline interface {
	OnDistEntityOnline(entity ec.Entity)
}

// EventDistEntityOffline 事件：分布式实体下线
type EventDistEntityOffline interface {
	OnDistEntityOffline(entity ec.Entity)
}

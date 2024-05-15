package gap

import (
	"git.golaxy.org/core/util/types"
	"hash/fnv"
	"reflect"
)

// MsgId 消息Id
type MsgId = uint32

// MakeMsgId 创建类型Id
func MakeMsgId(msg Msg) MsgId {
	hash := fnv.New32a()
	rt := reflect.ValueOf(msg).Elem().Type()
	if rt.PkgPath() == "" || rt.Name() == "" {
		panic("unsupported type")
	}
	hash.Write([]byte(types.TypeFullName(rt)))
	return MsgId(MsgId_Customize + hash.Sum32())
}

// MakeMsgIdT 创建类型Id
func MakeMsgIdT[T any]() MsgId {
	hash := fnv.New32a()
	rt := reflect.ValueOf((*T)(nil)).Elem().Type()
	if rt.PkgPath() == "" || rt.Name() == "" || !reflect.PointerTo(rt).Implements(reflect.TypeFor[Msg]()) {
		panic("unsupported type")
	}
	hash.Write([]byte(types.TypeFullName(rt)))
	return MsgId(MsgId_Customize + hash.Sum32())
}

package gap

import "io"

type TypeId = uint32

const (
	TypeId_None TypeId = iota
	TypeId_Int
	TypeId_Int8
	TypeId_Int32
	TypeId_Int64
	TypeId_Uint
	TypeId_Uint8
	TypeId_Uint32
	TypeId_Uint64
	TypeId_Byte
	TypeId_Bool
	TypeId_Bytes
	TypeId_String
)

type Var interface {
	io.ReadWriter
	// Size 大小
	Size() int
	//
	Type() TypeId
}

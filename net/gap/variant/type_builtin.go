package variant

const (
	TypeId_None TypeId = iota
	TypeId_Int
	TypeId_Int8
	TypeId_Int16
	TypeId_Int32
	TypeId_Int64
	TypeId_Uint
	TypeId_Uint8
	TypeId_Uint16
	TypeId_Uint32
	TypeId_Uint64
	TypeId_Float
	TypeId_Double
	TypeId_Byte
	TypeId_Bool
	TypeId_Bytes
	TypeId_String
	TypeId_Null
	TypeId_Array
	TypeId_Map
	TypeId_Error
	TypeId_Customize = 32 // 自定义类型起点
)

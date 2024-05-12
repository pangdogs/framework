package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// String string
type String string

// Read implements io.Reader
func (v String) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteString(string(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *String) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = String(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (v String) Size() int {
	return binaryutil.SizeofString(string(v))
}

// TypeId 类型
func (String) TypeId() TypeId {
	return TypeId_String
}

// Indirect 原始值
func (v String) Indirect() any {
	return string(v)
}

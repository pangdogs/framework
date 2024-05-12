package variant

import (
	"git.golaxy.org/framework/util/binaryutil"
)

// Double builtin double
type Double float64

// Read implements io.Reader
func (v Double) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteDouble(float64(v)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Double) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	val, err := bs.ReadDouble()
	if err != nil {
		return bs.BytesRead(), err
	}
	*v = Double(val)
	return bs.BytesRead(), nil
}

// Size 大小
func (Double) Size() int {
	return binaryutil.SizeofDouble()
}

// TypeId 类型
func (Double) TypeId() TypeId {
	return TypeId_Double
}

// Indirect 原始值
func (v Double) Indirect() any {
	return float64(v)
}

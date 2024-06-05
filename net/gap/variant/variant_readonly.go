package variant

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/util/binaryutil"
)

// VariantReadonly 只读可变类型
type VariantReadonly struct {
	TypeId TypeId      // 类型Id
	Value  ValueReader // 只读值
}

// Read implements io.Reader
func (v VariantReadonly) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.ReadFrom(&bs, v.TypeId); err != nil {
		return bs.BytesWritten(), err
	}

	if v.Value == nil {
		return bs.BytesWritten(), errors.New("value is nil")
	}

	if _, err := binaryutil.ReadFrom(&bs, v.Value); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), nil
}

// Size 大小
func (v VariantReadonly) Size() int {
	n := v.TypeId.Size()
	if v.Value != nil {
		n += v.Value.Size()
	}
	return n
}

// MakeVariantReadonly 创建只读可变类型
func MakeVariantReadonly(v ValueReader) (VariantReadonly, error) {
	if v == nil {
		return VariantReadonly{}, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}
	return VariantReadonly{
		TypeId: v.TypeId(),
		Value:  v,
	}, nil
}

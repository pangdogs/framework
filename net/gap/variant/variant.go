package variant

import (
	"errors"
	"git.golaxy.org/framework/util/binaryutil"
	"reflect"
)

// Variant 可变类型
type Variant struct {
	TypeId    TypeId        // 类型Id
	Value     ValueReader   // 读取值
	Reflected reflect.Value // 反射值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
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

// Write implements io.Writer
func (v *Variant) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&v.TypeId); err != nil {
		return bs.BytesRead(), err
	}

	reflected, err := v.TypeId.NewReflected()
	if err != nil {
		return bs.BytesRead(), err
	}

	value := reflected.Interface().(Value)
	if _, err := bs.WriteTo(value); err != nil {
		return bs.BytesRead(), err
	}

	v.Value = value
	v.Reflected = reflected

	return bs.BytesRead(), nil
}

// Size 大小
func (v Variant) Size() int {
	n := v.TypeId.Size()
	if v.Value != nil {
		n += v.Value.Size()
	}
	return n
}

// MakeVariant 创建可变类型
func MakeVariant(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, errors.New("v is nil")
	}
	return Variant{
		TypeId: v.TypeId(),
		Value:  v,
	}, nil
}

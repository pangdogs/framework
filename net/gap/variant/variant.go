package variant

import (
	"errors"
	"git.golaxy.org/framework/util/binaryutil"
	"reflect"
)

// Variant 可变类型
type Variant struct {
	TypeId        TypeId        // 类型Id
	ValueReadonly ValueReader   // 只读值
	Value         Value         // 值
	Reflected     reflect.Value // 反射值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
	if !v.Valid() {
		return 0, errors.New("variant is invalid")
	}

	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.ReadFrom(&bs, v.TypeId); err != nil {
		return bs.BytesWritten(), err
	}

	if v.Readonly() {
		if _, err := binaryutil.ReadFrom(&bs, v.ValueReadonly); err != nil {
			return bs.BytesWritten(), err
		}
	} else {
		if _, err := binaryutil.ReadFrom(&bs, v.Value); err != nil {
			return bs.BytesWritten(), err
		}
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

	v.ValueReadonly = nil
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

// Readonly 只读
func (v Variant) Readonly() bool {
	return v.ValueReadonly != nil
}

// Valid 有效
func (v Variant) Valid() bool {
	if v.Readonly() {
		return v.TypeId == v.ValueReadonly.TypeId()
	}
	if v.Value != nil {
		return v.TypeId == v.Value.TypeId()
	}
	return false
}

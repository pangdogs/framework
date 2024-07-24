package variant

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

// MakeSerializedValue 创建已序列化值
func MakeSerializedValue(v ValueReader) (ret *SerializedValue, err error) {
	if v == nil {
		return nil, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}

	sv := &SerializedValue{
		Type: v.TypeId(),
	}

	size := v.Size()
	if size > 0 {
		buff := binaryutil.MakeRecycleBytes(size)
		defer func() {
			if ret == nil {
				buff.Release()
			}
		}()

		if _, err := v.Read(buff.Data()); err != nil {
			return nil, err
		}

		sv.Data = buff

	} else {
		sv.Data = binaryutil.NilRecycleBytes
	}

	return sv, nil
}

// SerializedValue 已序列化值
type SerializedValue struct {
	Type TypeId                  // 类型Id
	Data binaryutil.RecycleBytes // 数据
}

// Read implements io.Reader
func (v *SerializedValue) Read(p []byte) (int, error) {
	if len(p) < len(v.Data.Data()) {
		return 0, io.ErrShortWrite
	}
	return copy(p, v.Data.Data()), nil
}

// Size 大小
func (v *SerializedValue) Size() int {
	return len(v.Data.Data())
}

// TypeId 类型
func (v *SerializedValue) TypeId() TypeId {
	return v.Type
}

// Indirect 原始值
func (v *SerializedValue) Indirect() any {
	return v
}

// Release 释放资源
func (v *SerializedValue) Release() {
	v.Data.Release()
}

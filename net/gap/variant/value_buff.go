package variant

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/util/binaryutil"
)

// MakeValueBuff 创建ValueBuff
func MakeValueBuff(v ValueReader) (*ValueBuff, error) {
	if v == nil {
		return nil, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}

	valueBuff := &ValueBuff{
		Type: v.TypeId(),
	}

	s := v.Size()
	if s > 0 {
		buff := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(s))

		if _, err := v.Read(buff.Data()); err != nil {
			buff.Release()
			return nil, err
		}

		valueBuff.Buff = buff

	} else {
		valueBuff.Buff = binaryutil.NilRecycleBytes
	}

	return valueBuff, nil
}

// ValueBuff ValueBuff
type ValueBuff struct {
	Type TypeId
	Buff binaryutil.RecycleBytes
}

// Read implements io.Reader
func (v *ValueBuff) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteBytes(v.Buff.Data()); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), nil
}

// Size 大小
func (v *ValueBuff) Size() int {
	return len(v.Buff.Data())
}

// TypeId 类型
func (v *ValueBuff) TypeId() TypeId {
	return v.Type
}

// Indirect 原始值
func (v *ValueBuff) Indirect() any {
	return v
}

// Release 释放资源
func (v *ValueBuff) Release() {
	v.Buff.Release()
}

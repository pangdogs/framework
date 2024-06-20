package variant

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/utils/binaryutil"
)

// Map map
type Map generic.UnorderedSliceMap[Variant, Variant]

// Read implements io.Reader
func (v Map) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		kv := &v[i]

		if _, err := binaryutil.ReadFrom(&bs, kv.K); err != nil {
			return bs.BytesWritten(), err
		}

		if _, err := binaryutil.ReadFrom(&bs, kv.V); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Map) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]generic.UnorderedKV[Variant, Variant], l)

	for i := uint64(0); i < l; i++ {
		kv := &(*v)[i]

		if _, err := bs.WriteTo(&kv.K); err != nil {
			return bs.BytesRead(), err
		}

		if _, err := bs.WriteTo(&kv.V); err != nil {
			return bs.BytesRead(), err
		}
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v Map) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		kv := &v[i]
		n += kv.K.Size()
		n += kv.V.Size()
	}
	return n
}

// TypeId 类型
func (Map) TypeId() TypeId {
	return TypeId_Map
}

// Indirect 原始值
func (v Map) Indirect() any {
	return v
}

// Release 释放资源
func (v Map) Release() {
	for i := range v {
		kv := &v[i]

		if kv.K.Readonly() {
			kv.K.ValueReadonly.Release()
		}

		if kv.V.Readonly() {
			kv.V.ValueReadonly.Release()
		}
	}
}

// ToUnorderedSliceMap 转换为UnorderedSliceMap
func (v *Map) ToUnorderedSliceMap() *generic.UnorderedSliceMap[Variant, Variant] {
	return (*generic.UnorderedSliceMap[Variant, Variant])(v)
}

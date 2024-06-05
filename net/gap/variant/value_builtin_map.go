package variant

import (
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/util/binaryutil"
)

// MakeMapFromGoMap 创建可变类型map
func MakeMapFromGoMap[K comparable, V any](m map[K]V) (Map, error) {
	varMap := make(Map, 0, len(m))

	for k, v := range m {
		var rv generic.KV[Variant, Variant]
		var err error

		rv.K, err = CastVariant(k)
		if err != nil {
			return nil, err
		}

		rv.V, err = CastVariant(v)
		if err != nil {
			return nil, err
		}

		varMap = append(varMap, rv)
	}

	return varMap, nil
}

// MakeMapFromSliceMap 创建可变类型map
func MakeMapFromSliceMap[K comparable, V any](m generic.SliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for _, kv := range m {
		var rv generic.KV[Variant, Variant]
		var err error

		rv.K, err = CastVariant(kv.K)
		if err != nil {
			return nil, err
		}

		rv.V, err = CastVariant(kv.V)
		if err != nil {
			return nil, err
		}

		varMap = append(varMap, rv)
	}

	return varMap, nil
}

// Map map
type Map generic.SliceMap[Variant, Variant]

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

	*v = make([]generic.KV[Variant, Variant], l)

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

// CastSliceMap 转换为SliceMap
func (v Map) CastSliceMap() generic.SliceMap[Variant, Variant] {
	return (generic.SliceMap[Variant, Variant])(v)
}

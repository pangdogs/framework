package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MakeMap 创建可变类型map
func MakeMap[K comparable, V any](m map[K]V) (Map, error) {
	varMap := make(Map, 0, len(m))

	for k, v := range m {
		var kv KV
		var err error

		kv.K, err = CastVariant(k)
		if err != nil {
			return nil, err
		}

		kv.V, err = CastVariant(v)
		if err != nil {
			return nil, err
		}

		varMap = append(varMap, kv)
	}

	return varMap, nil
}

// KV kv
type KV struct {
	K, V Variant
}

// Map map
type Map []KV

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

	*v = make([]KV, l)

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

// Type 类型
func (Map) Type() TypeId {
	return TypeId_Map
}

// Indirect 原始值
func (v Map) Indirect() any {
	return v
}

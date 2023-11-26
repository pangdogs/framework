package variant

import (
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MakeArray 创建Array
func MakeArray[T any](arr []T) (Array, error) {
	var varArr Array

	for i := range arr {
		variant, err := CastVariant(arr[i])
		if err != nil {
			return nil, err
		}
		varArr = append(varArr, variant)
	}

	return varArr, nil
}

// Array 数组
type Array []Variant

// Read implements io.Reader
func (v Array) Read(p []byte) (int, error) {
	rn := 0

	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return rn, err
	}
	rn += bs.BytesWritten()

	for i := range v {
		n, err := v[i].Read(p[rn:])
		rn += n
		if err != nil {
			return rn, err
		}
	}

	return rn, nil
}

// Write implements io.Writer
func (v *Array) Write(p []byte) (int, error) {
	wn := 0

	bs := binaryutil.NewBigEndianStream(p)
	l, err := bs.ReadUvarint()
	if err != nil {
		return wn, err
	}
	wn += bs.BytesRead()

	arr := make([]Variant, l)

	for i := uint64(0); i < l; i++ {
		n, err := arr[i].Write(p[wn:])
		wn += n
		if err != nil {
			return wn, err
		}
	}
	*v = arr

	return wn, nil
}

// Size 大小
func (v Array) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += v[i].Size()
	}
	return n
}

// Type 类型
func (Array) Type() TypeId {
	return TypeId_Array
}

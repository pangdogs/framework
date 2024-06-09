package variant

import "git.golaxy.org/framework/util/binaryutil"

type Call struct {
	Src, Transit string
}

type CallChain []Call

// Read implements io.Reader
func (v CallChain) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		if err := bs.WriteString(v[i].Src); err != nil {
			return bs.BytesWritten(), err
		}
		if err := bs.WriteString(v[i].Transit); err != nil {
			return bs.BytesWritten(), err
		}
	}

	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *CallChain) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	l, err := bs.ReadUvarint()
	if err != nil {
		return bs.BytesRead(), err
	}

	*v = make([]Call, l)

	for i := uint64(0); i < l; i++ {
		src, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}

		transit, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}

		(*v)[i].Src = src
		(*v)[i].Transit = transit
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v CallChain) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += binaryutil.SizeofString(v[i].Src)
		n += binaryutil.SizeofString(v[i].Transit)
	}
	return n
}

// TypeId 类型
func (CallChain) TypeId() TypeId {
	return TypeId_CallChain
}

// Indirect 原始值
func (v CallChain) Indirect() any {
	return v
}

// Release 释放资源
func (CallChain) Release() {}

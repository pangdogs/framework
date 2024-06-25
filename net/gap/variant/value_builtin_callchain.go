package variant

import "git.golaxy.org/framework/utils/binaryutil"

type Call struct {
	Svc, Addr, Transit string
}

type CallChain []Call

// Read implements io.Reader
func (v CallChain) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if err := bs.WriteUvarint(uint64(len(v))); err != nil {
		return bs.BytesWritten(), err
	}

	for i := range v {
		if err := bs.WriteString(v[i].Svc); err != nil {
			return bs.BytesWritten(), err
		}
		if err := bs.WriteString(v[i].Addr); err != nil {
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
		svc, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}

		addr, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}

		transit, err := bs.ReadString()
		if err != nil {
			return bs.BytesRead(), err
		}

		(*v)[i].Svc = svc
		(*v)[i].Addr = addr
		(*v)[i].Transit = transit
	}

	return bs.BytesRead(), nil
}

// Size 大小
func (v CallChain) Size() int {
	n := binaryutil.SizeofUvarint(uint64(len(v)))
	for i := range v {
		n += binaryutil.SizeofString(v[i].Svc)
		n += binaryutil.SizeofString(v[i].Addr)
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

func (v CallChain) First() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[0]
}

func (v CallChain) Last() Call {
	if len(v) <= 0 {
		return Call{}
	}
	return v[len(v)-1]
}

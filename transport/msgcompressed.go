package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgCompressed 压缩消息
type MsgCompressed struct {
	Data   []byte
	RawLen int64
}

func (m *MsgCompressed) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteBytes(p); err != nil {
		return 0, err
	}
	if err := bs.WriteVarint(m.RawLen); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgCompressed) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	rawLen, err := bs.ReadVarint()
	if err != nil {
		return 0, err
	}
	m.Data = data
	m.RawLen = rawLen
	return bs.BytesRead(), nil
}

func (m *MsgCompressed) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofVarint(m.RawLen)
}

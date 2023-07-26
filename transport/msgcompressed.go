package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

// MsgCompressed 压缩消息
type MsgCompressed struct {
	Data         []byte
	OriginalSize int64
}

// Read implements io.Reader
func (m *MsgCompressed) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteBytes(m.Data); err != nil {
		return 0, err
	}
	if err := bs.WriteVarint(m.OriginalSize); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (m *MsgCompressed) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	data, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	originalSize, err := bs.ReadVarint()
	if err != nil {
		return 0, err
	}
	m.Data = data
	m.OriginalSize = originalSize
	return bs.BytesRead(), nil
}

// Size 大小
func (m *MsgCompressed) Size() int {
	return binaryutil.SizeofBytes(m.Data) + binaryutil.SizeofVarint(m.OriginalSize)
}

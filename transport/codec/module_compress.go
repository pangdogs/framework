package codec

import (
	"bytes"
	"errors"
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/xio"
	"math"
)

// ICompressModule 压缩模块接口
type ICompressModule interface {
	// Compress 压缩数据
	Compress(src []byte) (dst []byte, compressed bool, err error)
	// Uncompress 解压缩数据
	Uncompress(src []byte) (dst []byte, err error)
	// GC GC
	GC()
}

// CompressModule 压缩模块
type CompressModule struct {
	NewReader func(io.Reader) (io.Reader, error)      // 解压缩函数
	NewWriter func(io.Writer) (io.WriteCloser, error) // 压缩函数
	gcList    [][]byte                                // GC列表
}

// Compress 压缩数据
func (m *CompressModule) Compress(src []byte) (dst []byte, compressed bool, err error) {
	if m.NewWriter == nil {
		return nil, false, errors.New("setting NewWriter is nil")
	}

	if len(src) <= 0 {
		return src, false, nil
	}

	compressedBuf := BytesPool.Get(len(src))
	defer BytesPool.Put(compressedBuf)

	n, err := func() (n int, err error) {
		lw := xio.NewBytesWriter(compressedBuf)
		w, err := m.NewWriter(lw)
		if err != nil {
			return 0, err
		}
		defer func() {
			if err == nil {
				if err = w.Close(); err == nil {
					n = lw.N
				}
			} else {
				w.Close()
			}
		}()

		_, err = w.Write(src)
		return
	}()
	if err != nil {
		if errors.Is(err, xio.ErrLimitReached) {
			return src, false, nil
		}
		return nil, false, err
	}

	msgCompressed := transport.MsgCompressed{
		Data:   compressedBuf[:n],
		RawLen: int64(len(src)),
	}

	if msgCompressed.Size() >= len(src) {
		return src, false, nil
	}

	buf := BytesPool.Get(msgCompressed.Size())
	defer func() {
		if compressed {
			m.gcList = append(m.gcList, buf)
		} else {
			BytesPool.Put(buf)
		}
	}()

	if _, err = msgCompressed.Read(buf); err != nil {
		return nil, false, err
	}

	return buf, true, nil
}

// Uncompress 解压缩数据
func (m *CompressModule) Uncompress(src []byte) (dst []byte, err error) {
	if m.NewReader == nil {
		return nil, errors.New("setting NewReader is nil")
	}

	if len(src) <= 0 {
		return nil, errors.New("src bytes too small")
	}

	msgCompressed := transport.MsgCompressed{}

	_, err = msgCompressed.Write(src)
	if err != nil {
		return nil, err
	}

	if msgCompressed.RawLen >= math.MaxInt32 {
		return nil, errors.New("raw len too large")
	}

	rawBuf := BytesPool.Get(int(msgCompressed.RawLen))
	defer func() {
		if err == nil {
			m.gcList = append(m.gcList, rawBuf)
		} else {
			BytesPool.Put(rawBuf)
		}
	}()

	r, err := m.NewReader(bytes.NewReader(msgCompressed.Data))
	if err != nil {
		return nil, err
	}

	if _, err = r.Read(rawBuf); !errors.Is(err, io.EOF) {
		return nil, err
	}

	return rawBuf, nil
}

// GC GC
func (m *CompressModule) GC() {
	for i := range m.gcList {
		BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

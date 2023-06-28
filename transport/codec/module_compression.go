package codec

import (
	"bytes"
	"errors"
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/method"
	"kit.golaxy.org/plugins/transport/xio"
	"math"
)

// ICompressionModule 压缩模块接口
type ICompressionModule interface {
	// Compress 压缩数据
	Compress(src []byte) (dst []byte, compressed bool, err error)
	// Uncompress 解压缩数据
	Uncompress(src []byte) (dst []byte, err error)
	// GC GC
	GC()
}

// CompressionModule 压缩模块
type CompressionModule struct {
	CompressionStream method.CompressionStream // 压缩流
	gcList            [][]byte                 // GC列表
}

// Compress 压缩数据
func (m *CompressionModule) Compress(src []byte) (dst []byte, compressed bool, err error) {
	if m.CompressionStream == nil {
		return nil, false, errors.New("setting CompressionStream is nil")
	}

	if len(src) <= 0 {
		return src, false, nil
	}

	compressedBuf := BytesPool.Get(len(src))
	defer BytesPool.Put(compressedBuf)

	n, err := func() (n int, err error) {
		bw := xio.NewBytesWriter(compressedBuf)
		w, err := m.CompressionStream.WrapWriter(bw)
		if err != nil {
			return 0, err
		}
		defer func() {
			if err == nil {
				if err = w.Close(); err == nil {
					n = bw.N
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
		Data:         compressedBuf[:n],
		OriginalSize: int64(len(src)),
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
func (m *CompressionModule) Uncompress(src []byte) (dst []byte, err error) {
	if m.CompressionStream == nil {
		return nil, errors.New("setting CompressionStream is nil")
	}

	if len(src) <= 0 {
		return nil, errors.New("src too small")
	}

	msgCompressed := transport.MsgCompressed{}

	_, err = msgCompressed.Write(src)
	if err != nil {
		return nil, err
	}

	if msgCompressed.OriginalSize >= math.MaxInt32 {
		return nil, errors.New("OriginalSize too large")
	}

	rawBuf := BytesPool.Get(int(msgCompressed.OriginalSize))
	defer func() {
		if err == nil {
			m.gcList = append(m.gcList, rawBuf)
		} else {
			BytesPool.Put(rawBuf)
		}
	}()

	r, err := m.CompressionStream.WrapReader(bytes.NewReader(msgCompressed.Data))
	if err != nil {
		return nil, err
	}

	if _, err = r.Read(rawBuf); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return rawBuf, nil
}

// GC GC
func (m *CompressionModule) GC() {
	for i := range m.gcList {
		BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

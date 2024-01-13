package codec

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/plugins/gtp"
	"git.golaxy.org/plugins/gtp/method"
	"git.golaxy.org/plugins/util/binaryutil"
	"io"
	"math"
)

// ICompressionModule 压缩模块接口
type ICompressionModule interface {
	// Compress 压缩数据
	Compress(src []byte) (dst binaryutil.RecycleBytes, compressed bool, err error)
	// Uncompress 解压缩数据
	Uncompress(src []byte) (dst binaryutil.RecycleBytes, err error)
}

// NewCompressionModule 创建压缩模块
func NewCompressionModule(cs method.CompressionStream) ICompressionModule {
	if cs == nil {
		panic(fmt.Errorf("%w: cs is nil", core.ErrArgs))
	}

	return &CompressionModule{
		CompressionStream: cs,
	}
}

// CompressionModule 压缩模块
type CompressionModule struct {
	CompressionStream method.CompressionStream // 压缩流
}

// Compress 压缩数据
func (m *CompressionModule) Compress(src []byte) (dst binaryutil.RecycleBytes, compressed bool, err error) {
	if len(src) <= 0 {
		return binaryutil.MakeNonRecycleBytes(src), false, nil
	}

	if m.CompressionStream == nil {
		return binaryutil.MakeNonRecycleBytes(nil), false, errors.New("setting CompressionStream is nil")
	}

	compressedBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(len(src)))
	defer compressedBuf.Release()

	n, err := func() (n int, err error) {
		bw := binaryutil.NewBytesWriter(compressedBuf.Data())
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
		if errors.Is(err, binaryutil.ErrLimitReached) {
			return binaryutil.MakeNonRecycleBytes(src), false, nil
		}
		return binaryutil.MakeNonRecycleBytes(nil), false, err
	}

	msgCompressed := gtp.MsgCompressed{
		Data:         compressedBuf.Data()[:n],
		OriginalSize: int64(len(src)),
	}

	if msgCompressed.Size() >= len(src) {
		return binaryutil.MakeNonRecycleBytes(src), false, nil
	}

	buf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msgCompressed.Size()))
	defer func() {
		if err != nil {
			buf.Release()
		}
	}()

	if _, err = msgCompressed.Read(buf.Data()); err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), false, err
	}

	return buf, true, nil
}

// Uncompress 解压缩数据
func (m *CompressionModule) Uncompress(src []byte) (dst binaryutil.RecycleBytes, err error) {
	if len(src) <= 0 {
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("%w: src too small", core.ErrArgs)
	}

	if m.CompressionStream == nil {
		return binaryutil.MakeNonRecycleBytes(nil), errors.New("setting CompressionStream is nil")
	}

	msgCompressed := gtp.MsgCompressed{}

	_, err = msgCompressed.Write(src)
	if err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), err
	}

	if msgCompressed.OriginalSize >= math.MaxInt32 {
		return binaryutil.MakeNonRecycleBytes(nil), errors.New("original size too large")
	}

	rawBuf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(int(msgCompressed.OriginalSize)))
	defer func() {
		if err != nil {
			rawBuf.Release()
		}
	}()

	r, err := m.CompressionStream.WrapReader(bytes.NewReader(msgCompressed.Data))
	if err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), err
	}

	if _, err = r.Read(rawBuf.Data()); err != nil && !errors.Is(err, io.EOF) {
		return binaryutil.MakeNonRecycleBytes(nil), err
	}

	return rawBuf, nil
}

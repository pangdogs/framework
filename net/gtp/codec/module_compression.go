/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package codec

import (
	"bytes"
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/utils/binaryutil"
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
		return binaryutil.NilRecycleBytes, false, errors.New("setting CompressionStream is nil")
	}

	compressedBuf := binaryutil.MakeRecycleBytes(len(src))
	defer compressedBuf.Release()

	n, err := func() (n int, err error) {
		bw := binaryutil.NewBytesWriter(compressedBuf.Data())
		w, err := m.CompressionStream.WrapWriter(bw)
		if err != nil {
			return 0, err
		}
		defer func() {
			closeErr := w.Close()
			if err == nil {
				err = closeErr
			}
			if err == nil {
				n = bw.N
			}
		}()

		_, err = w.Write(src)
		return
	}()
	if err != nil {
		if errors.Is(err, binaryutil.ErrLimitReached) {
			return binaryutil.MakeNonRecycleBytes(src), false, nil
		}
		return binaryutil.NilRecycleBytes, false, err
	}

	msgCompressed := gtp.MsgCompressed{
		Data:         compressedBuf.Data()[:n],
		OriginalSize: int64(len(src)),
	}

	if msgCompressed.Size() >= len(src) {
		return binaryutil.MakeNonRecycleBytes(src), false, nil
	}

	buf := binaryutil.MakeRecycleBytes(msgCompressed.Size())
	defer func() {
		if !buf.Equal(dst) {
			buf.Release()
		}
	}()

	if _, err = msgCompressed.Read(buf.Data()); err != nil {
		return binaryutil.NilRecycleBytes, false, err
	}

	return buf, true, nil
}

// Uncompress 解压缩数据
func (m *CompressionModule) Uncompress(src []byte) (dst binaryutil.RecycleBytes, err error) {
	if len(src) <= 0 {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: src too small", core.ErrArgs)
	}

	if m.CompressionStream == nil {
		return binaryutil.NilRecycleBytes, errors.New("setting CompressionStream is nil")
	}

	msgCompressed := gtp.MsgCompressed{}

	if _, err = msgCompressed.Write(src); err != nil {
		return binaryutil.NilRecycleBytes, err
	}

	if msgCompressed.OriginalSize >= math.MaxInt32 {
		return binaryutil.NilRecycleBytes, errors.New("original size too large")
	}

	buf := binaryutil.MakeRecycleBytes(int(msgCompressed.OriginalSize))
	defer func() {
		if !buf.Equal(dst) {
			buf.Release()
		}
	}()

	r, err := m.CompressionStream.WrapReader(bytes.NewReader(msgCompressed.Data))
	if err != nil {
		return binaryutil.NilRecycleBytes, err
	}

	if _, err = r.Read(buf.Data()); err != nil && !errors.Is(err, io.EOF) {
		return binaryutil.NilRecycleBytes, err
	}

	return buf, nil
}

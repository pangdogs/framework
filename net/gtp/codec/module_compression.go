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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	ErrCompress = errors.New("gtp-compress") // 压缩错误
)

// ICompression 压缩模块接口
type ICompression interface {
	// Compress 压缩数据
	Compress(src []byte) (compressedBuf binaryutil.Bytes, compressed bool, err error)
	// Uncompress 解压缩数据
	Uncompress(src []byte, max int) (uncompressedBuf binaryutil.Bytes, err error)
}

// NewCompression 创建压缩模块
func NewCompression(cs method.CompressionStream) ICompression {
	if cs == nil {
		exception.Panicf("%w: %w: cs is nil", ErrCompress, core.ErrArgs)
	}

	return &Compression{
		CompressionStream: cs,
	}
}

// Compression 压缩模块
type Compression struct {
	CompressionStream method.CompressionStream // 压缩流
}

// Compress 压缩数据
func (c *Compression) Compress(src []byte) (binaryutil.Bytes, bool, error) {
	if len(src) <= 0 {
		return binaryutil.EmptyBytes, false, nil
	}

	if c.CompressionStream == nil {
		return binaryutil.EmptyBytes, false, fmt.Errorf("%w: CompressionStream is nil", ErrCompress)
	}

	compressedDataBuf := binaryutil.NewBytes(true, len(src))

	n, err := func() (int, error) {
		bw := binaryutil.NewBytesWriter(compressedDataBuf.Payload())
		w, err := c.CompressionStream.WrapWriter(bw)
		if err != nil {
			return 0, err
		}
		if _, err := w.Write(src); err != nil {
			return 0, err
		}
		if err := w.Close(); err != nil {
			return 0, err
		}
		return bw.N, nil
	}()
	if err != nil {
		compressedDataBuf.Release()
		if errors.Is(err, binaryutil.ErrLimitReached) {
			return binaryutil.EmptyBytes, false, nil
		}
		return binaryutil.EmptyBytes, false, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	msgCompressed := gtp.MsgCompressed{
		Data:         compressedDataBuf.Payload()[:n],
		OriginalSize: int64(len(src)),
	}

	if msgCompressed.Size() >= len(src) {
		compressedDataBuf.Release()
		return binaryutil.EmptyBytes, false, nil
	}

	compressedBuf := binaryutil.NewBytes(true, msgCompressed.Size())

	if _, err := binaryutil.CopyToBuff(compressedBuf.Payload(), msgCompressed); err != nil {
		compressedBuf.Release()
		compressedDataBuf.Release()
		return binaryutil.EmptyBytes, false, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	compressedDataBuf.Release()
	return compressedBuf, true, nil
}

// Uncompress 解压缩数据
func (c *Compression) Uncompress(src []byte, max int) (binaryutil.Bytes, error) {
	if len(src) <= 0 {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w: src too small", ErrCompress, core.ErrArgs)
	}

	if c.CompressionStream == nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: CompressionStream is nil", ErrCompress)
	}

	msgCompressed := gtp.MsgCompressed{}

	if _, err := msgCompressed.Write(src); err != nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	if msgCompressed.OriginalSize < 0 {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: negative original size", ErrCompress)
	}
	if msgCompressed.OriginalSize > int64(max) {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: original size too large", ErrCompress)
	}

	uncompressedBuf := binaryutil.NewBytes(true, int(msgCompressed.OriginalSize))

	r, err := c.CompressionStream.WrapReader(bytes.NewReader(msgCompressed.Data))
	if err != nil {
		uncompressedBuf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	if _, err := binaryutil.CopyToBuff(uncompressedBuf.Payload(), r); err != nil {
		uncompressedBuf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	return uncompressedBuf, nil
}

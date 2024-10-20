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
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	ErrEncode = errors.New("gtp-encode") // 编码错误
)

// IEncoder 消息包编码器接口
type IEncoder interface {
	// Encode 编码消息包
	Encode(flags gtp.Flags, msg gtp.MsgReader) (binaryutil.RecycleBytes, error)
}

// Encoder 消息包编码器
type Encoder struct {
	EncryptionModule  IEncryptionModule  // 加密模块
	MACModule         IMACModule         // MAC模块
	CompressionModule ICompressionModule // 压缩模块
	Encryption        bool               // 开启加密
	PatchMAC          bool               // 开启MAC
	CompressedSize    int                // 启用压缩阀值（字节），<=0表示不开启
}

// Encode 编码消息包
func (e *Encoder) Encode(flags gtp.Flags, msg gtp.MsgReader) (ret binaryutil.RecycleBytes, err error) {
	if msg == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w: msg is nil", ErrEncode, core.ErrArgs)
	}

	head := gtp.MsgHead{}
	head.MsgId = msg.MsgId()

	head.Flags = flags.Setd(gtp.Flag_Encrypted, false).
		Setd(gtp.Flag_MAC, false).
		Setd(gtp.Flag_Compressed, false)

	// 预估追加的数据大小，因为后续数据可能会被压缩，所以此为评估值，只要保证不会内存溢出即可
	msgAddition := 0

	if e.Encryption {
		if e.EncryptionModule == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: EncryptionModule is nil, msg can't be encrypted", ErrEncode)
		}
		encAddition, err := e.EncryptionModule.SizeOfAddition(msg.Size())
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: encrypt SizeOfAddition failed, %w", ErrEncode, err)
		}
		msgAddition += encAddition

		if e.PatchMAC {
			if e.MACModule == nil {
				return binaryutil.NilRecycleBytes, fmt.Errorf("%w: MACModule is nil, msg can't be patch MAC", ErrEncode)
			}
			msgAddition += e.MACModule.SizeofMAC(msg.Size() + encAddition)
		}
	}

	mpBuf := binaryutil.MakeRecycleBytes(head.Size() + msg.Size() + msgAddition)
	defer func() {
		if !mpBuf.Equal(ret) {
			mpBuf.Release()
		}
	}()

	// 写入消息
	mn, err := binaryutil.ReadToBuff(mpBuf.Data()[head.Size():], msg)
	if err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: write msg failed, %w", ErrEncode, err)
	}
	end := head.Size() + int(mn)

	// 消息长度达到阀值，需要压缩消息
	if e.CompressedSize > 0 && msg.Size() >= e.CompressedSize {
		if e.CompressionModule == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: CompressionModule is nil, msg can't be compress", ErrEncode)
		}
		compressedBuf, compressed, err := e.CompressionModule.Compress(mpBuf.Data()[head.Size():end])
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: compress msg failed, %w", ErrEncode, err)
		}
		defer compressedBuf.Release()
		if compressed {
			head.Flags.Set(gtp.Flag_Compressed, true)

			copy(mpBuf.Data()[head.Size():], compressedBuf.Data())
			end = head.Size() + len(compressedBuf.Data())
		}
	}

	// 加密消息
	if e.Encryption {
		head.Flags.Set(gtp.Flag_Encrypted, true)

		// 补充MAC
		if e.PatchMAC {
			head.Flags.Set(gtp.Flag_MAC, true)

			if _, err = binaryutil.ReadToBuff(mpBuf.Data(), head); err != nil {
				return binaryutil.NilRecycleBytes, fmt.Errorf("%w: failed to write msg-packet-head for patch msg-mac, %w", ErrEncode, err)
			}

			macBuf, err := e.MACModule.PatchMAC(head.MsgId, head.Flags, mpBuf.Data()[head.Size():end])
			if err != nil {
				return binaryutil.NilRecycleBytes, fmt.Errorf("%w: patch msg-mac failed, %w", ErrEncode, err)
			}
			defer macBuf.Release()

			copy(mpBuf.Data()[head.Size():], macBuf.Data())
			end = head.Size() + len(macBuf.Data())
		}

		// 加密消息体
		encryptBuf, err := e.EncryptionModule.Transforming(mpBuf.Data()[head.Size():end], mpBuf.Data()[head.Size():end])
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: encrypt msg failed, %w", ErrEncode, err)
		}
		defer encryptBuf.Release()

		copy(mpBuf.Data()[head.Size():], encryptBuf.Data())
		end = head.Size() + len(encryptBuf.Data())
	}

	// 调整消息大小
	mpBuf = mpBuf.Slice(0, end)

	// 写入消息头
	head.Len = uint32(len(mpBuf.Data()))
	if _, err = binaryutil.ReadToBuff(mpBuf.Data(), head); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: write msg-packet-head failed, %w", ErrEncode, err)
	}

	return mpBuf, nil
}

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

// NewEncoder 创建消息包编码器
func NewEncoder() *Encoder {
	return &Encoder{}
}

// Encoder 消息包编码器
type Encoder struct {
	Encryption           IEncryption     // 加密模块
	Authentication       IAuthentication // 认证模块
	Compression          ICompression    // 压缩模块
	CompressionThreshold int             // 启用压缩阀值（字节），<=0表示不开启
}

// SetEncryption 设置加密模块
func (e *Encoder) SetEncryption(encryption IEncryption) *Encoder {
	e.Encryption = encryption
	return e
}

// SetAuthentication 设置认证模块
func (e *Encoder) SetAuthentication(authentication IAuthentication) *Encoder {
	e.Authentication = authentication
	return e
}

// SetCompression 设置压缩模块
func (e *Encoder) SetCompression(compression ICompression, compressionThreshold int) *Encoder {
	e.Compression = compression
	e.CompressionThreshold = compressionThreshold
	return e
}

// Encode 编码消息包
func (e *Encoder) Encode(flags gtp.Flags, msg gtp.ReadableMsg) (binaryutil.Bytes, error) {
	if msg == nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w: msg is nil", ErrEncode, core.ErrArgs)
	}

	head := gtp.MsgHead{}
	head.MsgId = msg.MsgId()

	head.Flags = flags.Setd(gtp.Flag_Encrypted, false).
		Setd(gtp.Flag_Signed, false).
		Setd(gtp.Flag_Compressed, false)

	// 预估追加的数据大小，因为后续数据可能会被压缩，所以此为评估值，只要保证不会内存溢出即可
	msgAddition := 0

	if e.Encryption != nil {
		encAddition, err := e.Encryption.SizeOfAddition(msg.Size())
		if err != nil {
			return binaryutil.EmptyBytes, fmt.Errorf("%w: encrypt SizeOfAddition failed, %w", ErrEncode, err)
		}
		msgAddition += encAddition

		if e.Authentication != nil {
			authAddition, err := e.Authentication.SizeOfAddition(msg.Size() + encAddition)
			if err != nil {
				return binaryutil.EmptyBytes, fmt.Errorf("%w: authenticate SizeOfAddition failed, %w", ErrEncode, err)
			}
			msgAddition += authAddition
		}
	}

	buf := binaryutil.NewBytes(true, head.Size()+msg.Size()+msgAddition)

	// 写入消息
	mn, err := binaryutil.CopyToBuff(buf.Payload()[head.Size():], msg)
	if err != nil {
		buf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: write msg failed, %w", ErrEncode, err)
	}
	end := head.Size() + int(mn)

	// 消息长度达到阀值，需要压缩消息
	if e.Compression != nil && e.CompressionThreshold > 0 && msg.Size() >= e.CompressionThreshold {
		compressedBuf, compressed, err := e.Compression.Compress(buf.Payload()[head.Size():end])
		if err != nil {
			buf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: compress msg failed, %w", ErrEncode, err)
		}
		if compressed {
			head.Flags.Set(gtp.Flag_Compressed, true)

			copy(buf.Payload()[head.Size():], compressedBuf.Payload())
			end = head.Size() + len(compressedBuf.Payload())

			compressedBuf.Release()
		}
	}

	// 加密消息
	if e.Encryption != nil {
		head.Flags.Set(gtp.Flag_Encrypted, true)

		// 消息签名
		if e.Authentication != nil {
			head.Flags.Set(gtp.Flag_Signed, true)

			if _, err = binaryutil.CopyToBuff(buf.Payload(), head); err != nil {
				buf.Release()
				return binaryutil.EmptyBytes, fmt.Errorf("%w: failed to write msg-packet-head for sign msg-mac, %w", ErrEncode, err)
			}

			macBuf, err := e.Authentication.Sign(head.MsgId, head.Flags, buf.Payload()[head.Size():end])
			if err != nil {
				buf.Release()
				return binaryutil.EmptyBytes, fmt.Errorf("%w: sign msg-mac failed, %w", ErrEncode, err)
			}

			copy(buf.Payload()[head.Size():], macBuf.Payload())
			end = head.Size() + len(macBuf.Payload())

			macBuf.Release()
		}

		// 加密消息体
		encryptBuf, err := e.Encryption.Transforming(buf.Payload()[head.Size():end], buf.Payload()[head.Size():end])
		if err != nil {
			buf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: encrypt msg failed, %w", ErrEncode, err)
		}

		copy(buf.Payload()[head.Size():], encryptBuf.Payload())
		end = head.Size() + len(encryptBuf.Payload())

		encryptBuf.Release()
	}

	// 调整消息大小
	buf = buf.Slice(0, end)

	// 写入消息头
	head.Len = uint32(len(buf.Payload()))
	if _, err = binaryutil.CopyToBuff(buf.Payload(), head); err != nil {
		buf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: write msg-packet-head failed, %w", ErrEncode, err)
	}

	return buf, nil
}

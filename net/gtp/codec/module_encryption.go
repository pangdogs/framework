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
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	// ErrEncrypt 是 GTP 消息加密和解密错误的根错误。
	ErrEncrypt = errors.New("gtp-encrypt")
)

// IEncryption 对 GTP 消息体执行加密或解密变换。
type IEncryption interface {
	// Transforming 将 src 变换到 dst 或新池化缓冲区；结果由调用方释放。
	Transforming(dst, src []byte) (transformedBuf binaryutil.Bytes, err error)
	// SizeOfAddition 返回变换结果相对输入可能增加的字节数。
	SizeOfAddition(msgLen int) (size int, err error)
}

type (
	// FetchNonce 为每次加密或解密变换提供 nonce。
	FetchNonce = generic.FuncPair0[[]byte, error]
)

// NewEncryption 创建加密或解密模块，并校验密码所需的填充器和 nonce 来源。
func NewEncryption(cipher method.Cipher, padding method.Padding, fetchNonce FetchNonce) IEncryption {
	if cipher == nil {
		exception.Panicf("%w: %w: cipher is nil", ErrEncrypt, core.ErrArgs)
	}

	if cipher.Pad() || cipher.Unpad() {
		if padding == nil {
			exception.Panicf("%w: %w: padding is nil", ErrEncrypt, core.ErrArgs)
		}
	}

	if cipher.NonceSize() > 0 {
		if fetchNonce == nil {
			exception.Panicf("%w: %w: fetchNonce is nil", ErrEncrypt, core.ErrArgs)
		}
	}

	return &Encryption{
		Cipher:     cipher,
		Padding:    padding,
		FetchNonce: fetchNonce,
	}
}

// Encryption 组合对称密码、可选填充和 nonce 来源。
type Encryption struct {
	Cipher     method.Cipher  // 加密或解密变换。
	Padding    method.Padding // 密码要求时使用的填充方案。
	FetchNonce FetchNonce     // 密码要求时为每次变换提供 nonce。
}

// Transforming 准备输入、获取 nonce、执行密码变换并按需填充或去除填充。
func (e *Encryption) Transforming(dst, src []byte) (binaryutil.Bytes, error) {
	if e.Cipher == nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: Cipher is nil", ErrEncrypt)
	}

	var inBuf binaryutil.Bytes

	inSize := e.Cipher.InputSize(len(src))
	if inSize > len(src) {
		inBuf = binaryutil.NewBytes(true, inSize)
		copy(inBuf.Payload(), src)
	} else {
		inBuf = binaryutil.RefBytes(src)
	}

	var outBuf binaryutil.Bytes

	outSize := e.Cipher.OutputSize(len(src))
	if outSize > len(dst) {
		outBuf = binaryutil.NewBytes(true, outSize)
	} else {
		outBuf = binaryutil.RefBytes(dst)
	}

	if e.Cipher.Pad() {
		if e.Padding == nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		if err := e.Padding.Pad(inBuf.Payload(), len(src)); err != nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	var nonce []byte

	if e.Cipher.NonceSize() > 0 {
		if e.FetchNonce == nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: FetchNonce is nil", ErrEncrypt)
		}
		var err error
		nonce, err = generic.FuncPairError(e.FetchNonce.SafeCall())
		if err != nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	ts, err := e.Cipher.Transforming(outBuf.Payload(), inBuf.Payload(), nonce)
	if err != nil {
		outBuf.Release()
		inBuf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
	}
	outBuf = outBuf.Slice(0, ts)

	if e.Cipher.Unpad() {
		if e.Padding == nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		buf, err := e.Padding.Unpad(outBuf.Payload())
		if err != nil {
			outBuf.Release()
			inBuf.Release()
			return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
		outBuf = outBuf.Slice(0, len(buf))
	}

	inBuf.Release()
	return outBuf, nil
}

// SizeOfAddition 返回当前密码变换最多增加的字节数。
func (e *Encryption) SizeOfAddition(msgLen int) (int, error) {
	if e.Cipher == nil {
		return 0, fmt.Errorf("%w: Cipher is nil", ErrEncrypt)
	}
	size := e.Cipher.OutputSize(msgLen) - msgLen
	if size < 0 {
		return 0, nil
	}
	return size, nil
}

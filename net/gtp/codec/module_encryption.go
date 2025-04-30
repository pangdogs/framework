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
	ErrEncrypt = errors.New("gtp-encrypt") // 加密错误
)

// IEncryption 加密模块接口
type IEncryption interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) (binaryutil.RecycleBytes, error)
	// SizeOfAddition 附加数据大小
	SizeOfAddition(msgLen int) (int, error)
}

type (
	FetchNonce = generic.FuncPair0[[]byte, error] // 获取nonce值
)

// NewEncryption 创建加密模块
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

// Encryption 加密模块
type Encryption struct {
	Cipher     method.Cipher  // 对称密码算法
	Padding    method.Padding // 填充方案
	FetchNonce FetchNonce     // 获取nonce值
}

// Transforming 变换数据
func (e *Encryption) Transforming(dst, src []byte) (ret binaryutil.RecycleBytes, err error) {
	if e.Cipher == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Cipher is nil", ErrEncrypt)
	}

	var in []byte

	is := e.Cipher.InputSize(len(src))
	if is > len(src) {
		buf := binaryutil.MakeRecycleBytes(is)
		defer buf.Release()

		copy(buf.Data(), src)
		in = buf.Data()
	} else {
		in = src
	}

	os := e.Cipher.OutputSize(len(src))
	if os > len(dst) {
		buf := binaryutil.MakeRecycleBytes(os)
		defer func() {
			if !buf.Equal(ret) {
				buf.Release()
			}
		}()

		ret = buf
	} else {
		ret = binaryutil.MakeNonRecycleBytes(dst)
	}

	if e.Cipher.Pad() {
		if e.Padding == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		if err = e.Padding.Pad(in, len(src)); err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	var nonce []byte

	if e.Cipher.NonceSize() > 0 {
		if e.FetchNonce == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: FetchNonce is nil", ErrEncrypt)
		}
		nonce, err = generic.FuncPairError(e.FetchNonce.SafeCall())
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	ts, err := e.Cipher.Transforming(ret.Data(), in, nonce)
	if err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
	}
	ret = ret.Slice(0, ts)

	if e.Cipher.Unpad() {
		if e.Padding == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		buf, err := e.Padding.Unpad(ret.Data())
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
		ret = ret.Slice(0, len(buf))
	}

	return ret, nil
}

// SizeOfAddition 附加数据大小
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

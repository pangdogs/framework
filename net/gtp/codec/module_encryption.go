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

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) (binaryutil.RecycleBytes, error)
	// SizeOfAddition 附加数据大小
	SizeOfAddition(msgLen int) (int, error)
}

type (
	FetchNonce = generic.PairFunc0[[]byte, error] // 获取nonce值
)

// NewEncryptionModule 创建加密模块
func NewEncryptionModule(cipher method.Cipher, padding method.Padding, fetchNonce FetchNonce) IEncryptionModule {
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

	return &EncryptionModule{
		Cipher:     cipher,
		Padding:    padding,
		FetchNonce: fetchNonce,
	}
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	Cipher     method.Cipher  // 对称密码算法
	Padding    method.Padding // 填充方案
	FetchNonce FetchNonce     // 获取nonce值
}

// Transforming 变换数据
func (m *EncryptionModule) Transforming(dst, src []byte) (ret binaryutil.RecycleBytes, err error) {
	if m.Cipher == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Cipher is nil", ErrEncrypt)
	}

	var in []byte

	is := m.Cipher.InputSize(len(src))
	if is > len(src) {
		buf := binaryutil.MakeRecycleBytes(is)
		defer buf.Release()

		copy(buf.Data(), src)
		in = buf.Data()
	} else {
		in = src
	}

	os := m.Cipher.OutputSize(len(src))
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

	if m.Cipher.Pad() {
		if m.Padding == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		if err = m.Padding.Pad(in, len(src)); err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	var nonce []byte

	if m.Cipher.NonceSize() > 0 {
		if m.FetchNonce == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: FetchNonce is nil", ErrEncrypt)
		}
		nonce, err = generic.PairFuncError(m.FetchNonce.Invoke())
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
	}

	ts, err := m.Cipher.Transforming(ret.Data(), in, nonce)
	if err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
	}
	ret = ret.Slice(0, ts)

	if m.Cipher.Unpad() {
		if m.Padding == nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Padding is nil", ErrEncrypt)
		}
		buf, err := m.Padding.Unpad(ret.Data())
		if err != nil {
			return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrEncrypt, err)
		}
		ret = ret.Slice(0, len(buf))
	}

	return ret, nil
}

// SizeOfAddition 附加数据大小
func (m *EncryptionModule) SizeOfAddition(msgLen int) (int, error) {
	if m.Cipher == nil {
		return 0, fmt.Errorf("%w: Cipher is nil", ErrEncrypt)
	}
	size := m.Cipher.OutputSize(msgLen) - msgLen
	if size < 0 {
		return 0, nil
	}
	return size, nil
}

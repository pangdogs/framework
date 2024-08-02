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

package method

import (
	"errors"
	"git.golaxy.org/framework/net/gtp"
)

// Padding 填充方案
type Padding interface {
	// Pad 填充
	Pad(buf []byte, ori int) error
	// Unpad 解除填充
	Unpad(padded []byte) ([]byte, error)
}

// NewPadding 创建填充方案
func NewPadding(pm gtp.PaddingMode) (Padding, error) {
	switch pm {
	case gtp.PaddingMode_Pkcs7:
		return _Pkcs7{}, nil
	case gtp.PaddingMode_X923:
		return _X923{}, nil
	default:
		return nil, ErrInvalidMethod
	}
}

type _Pkcs7 struct{}

// Pad 填充
func (_Pkcs7) Pad(buf []byte, ori int) error {
	padLen := len(buf) - ori
	if padLen <= 0 || padLen > 0xff {
		return errors.New("pkcs7: wrong pad length")
	}
	for i := 0; i < padLen; i++ {
		buf[ori+i] = byte(padLen)
	}
	return nil
}

// Unpad 解除填充
func (_Pkcs7) Unpad(padded []byte) ([]byte, error) {
	padLen := padded[len(padded)-1]
	padPos := len(padded) - int(padLen)
	if padPos < 0 {
		return nil, errors.New("pkcs7: wrong pad pos")
	}

	for i := len(padded) - 1; i >= padPos; i-- {
		if padded[i] != padLen {
			return nil, errors.New("pkcs7: incorrect padded")
		}
	}

	return padded[:padPos], nil
}

type _X923 struct{}

// Pad 填充
func (_X923) Pad(buf []byte, ori int) error {
	padLen := len(buf) - ori
	if padLen <= 0 || padLen > 0xff {
		return errors.New("x923: wrong pad length")
	}
	for i := 0; i < padLen-1; i++ {
		buf[ori+i] = 0
	}
	buf[ori+padLen-1] = byte(padLen)
	return nil
}

// Unpad 解除填充
func (_X923) Unpad(padded []byte) ([]byte, error) {
	padLen := padded[len(padded)-1]
	padPos := len(padded) - int(padLen)
	if padPos < 0 {
		return nil, errors.New("x923: wrong pad pos")
	}

	for i := len(padded) - 2; i >= padPos; i-- {
		if padded[i] != 0 {
			return nil, errors.New("x923: incorrect padded")
		}
	}

	return padded[:padPos], nil
}

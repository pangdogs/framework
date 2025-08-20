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
	"git.golaxy.org/framework/utils/binaryutil"
	"hash"
)

var (
	ErrAuthenticate = errors.New("gtp-authenticate")                 // 认证错误
	ErrInvalidMAC   = fmt.Errorf("%w: invalid MAC", ErrAuthenticate) // 错误的MAC
)

// IAuthentication 认证模块接口，用于防止消息被篡改
type IAuthentication interface {
	// Sign 签名
	Sign(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst binaryutil.RecycleBytes, err error)
	// Auth 认证
	Auth(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst []byte, err error)
	// SizeOfAddition 附加数据大小
	SizeOfAddition(msgLen int) (int, error)
}

// NewAuthentication 创建认证模块
func NewAuthentication(hmac hash.Hash) IAuthentication {
	if hmac == nil {
		exception.Panicf("%w: %w: HMAC is nil", ErrAuthenticate, core.ErrArgs)
	}

	return &Authentication{
		HMAC: hmac,
	}
}

// Authentication 认证模块
type Authentication struct {
	HMAC    hash.Hash
	macBuff []byte
}

// Sign 签名
func (m *Authentication) Sign(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst binaryutil.RecycleBytes, err error) {
	if m.HMAC == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}

	if len(m.macBuff) <= 0 {
		m.macBuff = make([]byte, m.HMAC.Size())
	}

	m.HMAC.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.HMAC.Write(bs[:])
	m.HMAC.Write(msgBuf)

	msgSigned := gtp.MsgSigned{
		Data: msgBuf,
		MAC:  m.HMAC.Sum(m.macBuff[:0]),
	}

	buf := binaryutil.MakeRecycleBytes(msgSigned.Size())
	defer func() {
		if !buf.Equal(dst) {
			buf.Release()
		}
	}()

	if _, err = binaryutil.CopyToBuff(buf.Data(), msgSigned); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrAuthenticate, err)
	}

	return buf, nil
}

// Auth 认证
func (m *Authentication) Auth(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.HMAC == nil {
		return nil, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}

	if len(m.macBuff) <= 0 {
		m.macBuff = make([]byte, m.HMAC.Size())
	}

	msgSigned := gtp.MsgSigned{}

	if _, err = msgSigned.Write(msgBuf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAuthenticate, err)
	}

	m.HMAC.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.HMAC.Write(bs[:])
	m.HMAC.Write(msgSigned.Data)

	if bytes.Compare(m.HMAC.Sum(m.macBuff[:0]), msgSigned.MAC) != 0 {
		return nil, ErrInvalidMAC
	}

	return msgSigned.Data, nil
}

// SizeOfAddition 附加数据大小
func (m *Authentication) SizeOfAddition(msgLen int) (int, error) {
	if m.HMAC == nil {
		return 0, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}
	return binaryutil.SizeofVarint(int64(msgLen)) + binaryutil.SizeofVarint(int64(m.HMAC.Size())) + m.HMAC.Size(), nil
}

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
	"crypto/hmac"
	"errors"
	"fmt"
	"hash"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	// ErrAuthenticate 是 GTP 消息认证错误的根错误。
	ErrAuthenticate = errors.New("gtp-authenticate")
	// ErrInvalidMAC 表示消息认证码校验失败。
	ErrInvalidMAC = fmt.Errorf("%w: invalid MAC", ErrAuthenticate)
)

// IAuthentication 为消息类型、标志位和消息体生成并验证认证码。
type IAuthentication interface {
	// Sign 包装消息体和认证码；返回的池化缓冲区由调用方释放。
	Sign(msgID gtp.MsgID, flags gtp.Flags, msgBuf []byte) (signedBuf binaryutil.Bytes, err error)
	// Auth 验证认证码并返回包装中的原始消息体。
	Auth(msgID gtp.MsgID, flags gtp.Flags, msgBuf []byte) (authBuf []byte, err error)
	// SizeOfAddition 返回认证包装相对原消息体增加的字节数。
	SizeOfAddition(msgLen int) (size int, err error)
}

// NewAuthentication 创建使用指定 HMAC 的消息认证模块；hmac 不得为 nil。
func NewAuthentication(hmac hash.Hash) IAuthentication {
	if hmac == nil {
		exception.Panicf("%w: %w: HMAC is nil", ErrAuthenticate, core.ErrArgs)
	}

	return &Authentication{
		HMAC: hmac,
	}
}

// Authentication 使用可复用的 HMAC 状态认证消息，不支持并发调用。
type Authentication struct {
	HMAC      hash.Hash // 消息认证使用的 HMAC。
	hmacCache []byte
}

// Sign 计算消息认证码并返回池化的 MsgSigned 编码。
func (m *Authentication) Sign(msgID gtp.MsgID, flags gtp.Flags, msgBuf []byte) (binaryutil.Bytes, error) {
	if m.HMAC == nil {
		return binaryutil.EmptyBytes, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}

	if len(m.hmacCache) <= 0 {
		m.hmacCache = make([]byte, m.HMAC.Size())
	}

	m.HMAC.Reset()
	bs := [2]byte{msgID, byte(flags)}
	m.HMAC.Write(bs[:])
	m.HMAC.Write(msgBuf)

	msgSigned := gtp.MsgSigned{
		Data: msgBuf,
		MAC:  m.HMAC.Sum(m.hmacCache[:0]),
	}

	signedBuf := binaryutil.NewBytes(true, msgSigned.Size())

	if _, err := binaryutil.CopyToBuff(signedBuf.Payload(), msgSigned); err != nil {
		signedBuf.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: %w", ErrAuthenticate, err)
	}

	return signedBuf, nil
}

// Auth 验证 MsgSigned 中的认证码，并返回引用 msgBuf 的原始消息体。
func (m *Authentication) Auth(msgID gtp.MsgID, flags gtp.Flags, msgBuf []byte) ([]byte, error) {
	if m.HMAC == nil {
		return nil, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}

	if len(m.hmacCache) <= 0 {
		m.hmacCache = make([]byte, m.HMAC.Size())
	}

	msgSigned := gtp.MsgSigned{}

	if _, err := msgSigned.Write(msgBuf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAuthenticate, err)
	}

	m.HMAC.Reset()
	bs := [2]byte{msgID, byte(flags)}
	m.HMAC.Write(bs[:])
	m.HMAC.Write(msgSigned.Data)

	if !hmac.Equal(m.HMAC.Sum(m.hmacCache[:0]), msgSigned.MAC) {
		return nil, ErrInvalidMAC
	}

	return msgSigned.Data, nil
}

// SizeOfAddition 返回 MsgSigned 编码相对原消息体增加的字节数。
func (m *Authentication) SizeOfAddition(msgLen int) (int, error) {
	if m.HMAC == nil {
		return 0, fmt.Errorf("%w: HMAC is nil", ErrAuthenticate)
	}
	return binaryutil.SizeofVarint(int64(msgLen)) + binaryutil.SizeofVarint(int64(m.HMAC.Size())) + m.HMAC.Size(), nil
}

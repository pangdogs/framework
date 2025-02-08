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
	ErrMAC          = errors.New("gtp-mac")                   // MAC错误
	ErrIncorrectMAC = fmt.Errorf("%w: incorrect MAC", ErrMAC) // MAC值不正确
)

// IMACModule MAC模块接口
type IMACModule interface {
	// PatchMAC 补充MAC
	PatchMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst binaryutil.RecycleBytes, err error)
	// VerifyMAC 验证MAC
	VerifyMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst []byte, err error)
	// SizeofMAC MAC大小
	SizeofMAC(msgLen int) int
}

// NewMACModule 创建MAC模块
func NewMACModule(h hash.Hash, pk []byte) IMACModule {
	if h == nil {
		exception.Panicf("%w: %w: h is nil", ErrMAC, core.ErrArgs)
	}

	if len(pk) <= 0 {
		exception.Panicf("%w: %w: len(pk) <= 0", ErrMAC, core.ErrArgs)
	}

	return &MACModule{
		Hash:       h,
		PrivateKey: pk,
	}
}

// MACModule MAC模块
type MACModule struct {
	Hash       hash.Hash // hash函数
	PrivateKey []byte    // 秘钥
}

// PatchMAC 补充MAC
func (m *MACModule) PatchMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst binaryutil.RecycleBytes, err error) {
	if m.Hash == nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: Hash is nil", ErrMAC)
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgBuf)
	m.Hash.Write(m.PrivateKey)

	msgMAC := gtp.MsgMAC{
		Data: msgBuf,
		MAC:  m.Hash.Sum(nil),
	}

	buf := binaryutil.MakeRecycleBytes(msgMAC.Size())
	defer func() {
		if !buf.Equal(dst) {
			buf.Release()
		}
	}()

	if _, err = binaryutil.CopyToBuff(buf.Data(), msgMAC); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: %w", ErrMAC, err)
	}

	return buf, nil
}

// VerifyMAC 验证MAC
func (m *MACModule) VerifyMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, fmt.Errorf("%w: Hash is nil", ErrMAC)
	}

	msgMAC := gtp.MsgMAC{}

	if _, err = msgMAC.Write(msgBuf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMAC, err)
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgMAC.Data)
	m.Hash.Write(m.PrivateKey)

	if bytes.Compare(m.Hash.Sum(nil), msgMAC.MAC) != 0 {
		return nil, ErrIncorrectMAC
	}

	return msgMAC.Data, nil
}

// SizeofMAC MAC大小
func (m *MACModule) SizeofMAC(msgLen int) int {
	return binaryutil.SizeofVarint(int64(msgLen)) + binaryutil.SizeofVarint(int64(m.Hash.Size())) + m.Hash.Size()
}

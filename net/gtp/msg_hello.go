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

package gtp

import (
	"bytes"
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"strings"
)

// Hello消息标志位
const (
	Flag_HelloDone  Flag = 1 << (iota + Flag_Customize) // Hello完成，在服务端返回的Hello消息中携带，表示初步认可客户端连接
	Flag_Encryption                                     // 开启加密（协议优先考虑性能，要求安全性请直接使用TLS加密链路），在服务端返回的Hello消息中携带，表示链路需要加密，需要执行秘钥交换流程
	Flag_Auth                                           // 开启鉴权（基于token鉴权），在服务端返回的Hello消息中携带，表示链路需要认证，需要执行鉴权流程
	Flag_Continue                                       // 断线重连
)

// CipherSuite 密码学套件
type CipherSuite struct {
	SecretKeyExchange   SecretKeyExchange   // 秘钥交换函数
	SymmetricEncryption SymmetricEncryption // 对称加密算法
	BlockCipherMode     BlockCipherMode     // 分组密码模式
	PaddingMode         PaddingMode         // 填充方案
	MACHash             Hash                // MAC摘要函数
}

// ParseCipherSuite 解析配置字串
func ParseCipherSuite(str string) (CipherSuite, error) {
	cs := CipherSuite{}
	var err error

	for i, s := range strings.Split(str, "-") {
		s = strings.ToLower(s)

		switch i {
		case 0:
			cs.SecretKeyExchange, err = ParseSecretKeyExchange(s)
			if err != nil {
				return CipherSuite{}, err
			}
		case 1:
			cs.SymmetricEncryption, err = ParseSymmetricEncryption(s)
			if err != nil {
				return CipherSuite{}, err
			}
		case 2:
			cs.BlockCipherMode, err = ParseBlockCipherMode(s)
			if err != nil {
				return CipherSuite{}, err
			}
		case 3:
			cs.PaddingMode, err = ParsePaddingMode(s)
			if err != nil {
				return CipherSuite{}, err
			}
		case 4:
			cs.MACHash, err = ParseHash(s)
			if err != nil {
				return CipherSuite{}, err
			}
		}
	}

	return cs, nil
}

// String implements fmt.Stringer
func (cs CipherSuite) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", cs.SecretKeyExchange, cs.SymmetricEncryption, cs.BlockCipherMode, cs.PaddingMode, cs.MACHash)
}

// Read implements io.Reader
func (cs CipherSuite) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint8(uint8(cs.SecretKeyExchange)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(cs.SymmetricEncryption)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(cs.BlockCipherMode)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(cs.PaddingMode)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(cs.MACHash)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (cs *CipherSuite) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	secretKeyExchange, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.SecretKeyExchange = SecretKeyExchange(secretKeyExchange)

	symmetricEncryption, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.SymmetricEncryption = SymmetricEncryption(symmetricEncryption)

	blockCipherMode, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.BlockCipherMode = BlockCipherMode(blockCipherMode)

	paddingMode, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.PaddingMode = PaddingMode(paddingMode)

	macHash, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.MACHash = Hash(macHash)

	return bs.BytesRead(), nil
}

// Size 大小
func (CipherSuite) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8() +
		binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}

// MsgHello Hello消息（注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgHello struct {
	Version     Version     // 协议版本
	SessionId   string      // 会话Id，如果客户端上传空值，服务端将会分配新会话，如果非空值，服务端将尝试查找会话，查找失败会重置链路
	Random      []byte      // 随机数，用于秘钥交换
	CipherSuite CipherSuite // 密码学套件，客户端提交的密码学套件建议，服务端可能不采纳，以服务端返回的为准，若客户端不支持，直接切断链路
	Compression Compression // 压缩函数，客户端提交的压缩函数建议，服务端可能不采纳，以服务端返回的为准，若客户端不支持，直接切断链路
}

// Read implements io.Reader
func (m *MsgHello) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint16(uint16(m.Version)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.SessionId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Random); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.CipherSuite); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(m.Compression)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write implements io.Writer
func (m *MsgHello) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	version, err := bs.ReadUint16()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Version = Version(version)

	m.SessionId, err = bs.ReadStringRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Random, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	_, err = bs.WriteTo(&m.CipherSuite)
	if err != nil {
		return bs.BytesRead(), err
	}

	compression, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Compression = Compression(compression)

	return bs.BytesRead(), nil
}

// Size 大小
func (m *MsgHello) Size() int {
	return binaryutil.SizeofUint16() + binaryutil.SizeofString(m.SessionId) + binaryutil.SizeofBytes(m.Random) +
		m.CipherSuite.Size() + binaryutil.SizeofUint8()
}

// MsgId 消息Id
func (*MsgHello) MsgId() MsgId {
	return MsgId_Hello
}

// Clone 克隆消息对象
func (m *MsgHello) Clone() Msg {
	return &MsgHello{
		Version:     m.Version,
		SessionId:   strings.Clone(m.SessionId),
		Random:      bytes.Clone(m.Random),
		CipherSuite: m.CipherSuite,
		Compression: m.Compression,
	}
}

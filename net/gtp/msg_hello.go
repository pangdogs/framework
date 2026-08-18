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
	"io"
	"strings"

	"git.golaxy.org/framework/utils/binaryutil"
)

const (
	// Flag_HelloDone 表示服务端已初步接受客户端 Hello。
	Flag_HelloDone Flag = 1 << (iota + Flag_Customize)
	// Flag_Encryption 表示后续握手需要执行密钥交换并启用加密。
	Flag_Encryption
	// Flag_Auth 表示后续握手需要执行令牌鉴权。
	Flag_Auth
	// Flag_Continue 表示客户端请求续接既有会话。
	Flag_Continue
)

// CipherSuite 组合 GTP 负载加密和认证所需的算法。
type CipherSuite struct {
	SecretKeyExchange   SecretKeyExchange   // 会话密钥交换算法。
	SymmetricEncryption SymmetricEncryption // 负载对称加密算法。
	BlockCipherMode     BlockCipherMode     // 分组密码工作模式。
	PaddingMode         PaddingMode         // 对称加密填充方案。
	HMAC                Hash                // 消息认证码使用的摘要算法。
}

// ParseCipherSuite 解析以连字符分隔的“密钥交换-加密-模式-填充-HMAC”配置；缺省项保持零值。
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
			cs.HMAC, err = ParseHash(s)
			if err != nil {
				return CipherSuite{}, err
			}
		}
	}

	return cs, nil
}

// String 返回以连字符分隔的密码套件配置。
func (cs CipherSuite) String() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", cs.SecretKeyExchange, cs.SymmetricEncryption, cs.BlockCipherMode, cs.PaddingMode, cs.HMAC)
}

// Read 将密码套件编码到 p。
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
	if err := bs.WriteUint8(uint8(cs.HMAC)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码密码套件。
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

	hmac, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	cs.HMAC = Hash(hmac)

	return bs.BytesRead(), nil
}

// Size 返回密码套件的固定编码字节数。
func (CipherSuite) Size() int {
	return binaryutil.SizeofUint8 + binaryutil.SizeofUint8 + binaryutil.SizeofUint8 +
		binaryutil.SizeofUint8 + binaryutil.SizeofUint8
}

// MsgHello 协商协议版本、会话、密码套件和压缩算法。
// 直接通过 Write 或 Unmarshal 解码时，SessionID 和 Random 会引用输入切片；
// 输入将被复用或修改时应先 Clone。Decoder.Decode 返回的消息不引用调用方输入。
type MsgHello struct {
	Version     Version     // 协议版本。
	SessionID   string      // 会话 ID；客户端留空时由服务端分配，非空时用于查找待续接会话。
	Random      []byte      // 密钥派生使用的随机数。
	CipherSuite CipherSuite // 客户端提议或服务端选定的密码套件。
	Compression Compression // 客户端提议或服务端选定的压缩算法。
}

// Read 将 Hello 消息编码到 p。
func (m MsgHello) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint16(uint16(m.Version)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.SessionID); err != nil {
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

// Write 从 p 解码 Hello 消息，引用型字段会引用 p。
func (m *MsgHello) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	version, err := bs.ReadUint16()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.Version = Version(version)

	m.SessionID, err = bs.ReadStringRef()
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

// Size 返回 Hello 消息编码后的字节数。
func (m MsgHello) Size() int {
	return binaryutil.SizeofUint16 + binaryutil.SizeofString(m.SessionID) + binaryutil.SizeofBytes(m.Random) +
		m.CipherSuite.Size() + binaryutil.SizeofUint8
}

// MsgID 返回 Hello 消息的内置类型 ID。
func (MsgHello) MsgID() MsgID {
	return MsgID_Hello
}

// Clone 深复制所有引用型字段。
func (m MsgHello) Clone() Msg {
	return &MsgHello{
		Version:     m.Version,
		SessionID:   strings.Clone(m.SessionID),
		Random:      bytes.Clone(m.Random),
		CipherSuite: m.CipherSuite,
		Compression: m.Compression,
	}
}

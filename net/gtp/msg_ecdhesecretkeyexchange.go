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
	// Flag_Signature 表示 ECDHE 密钥交换消息携带可供对端验证的签名。
	Flag_Signature Flag = 1 << (iota + Flag_Customize)
)

// SignatureAlgorithm 组合非对称算法、签名填充和摘要算法。
type SignatureAlgorithm struct {
	AsymmetricEncryption AsymmetricEncryption // 非对称签名算法。
	PaddingMode          PaddingMode          // 签名填充方案。
	Hash                 Hash                 // 摘要算法。
}

// ParseSignatureAlgorithm 解析以连字符分隔的“非对称算法-填充-摘要”配置；缺省项保持零值。
func ParseSignatureAlgorithm(str string) (SignatureAlgorithm, error) {
	sa := SignatureAlgorithm{}
	var err error

	for i, s := range strings.Split(str, "-") {
		s = strings.ToLower(s)

		switch i {
		case 0:
			sa.AsymmetricEncryption, err = ParseAsymmetricEncryption(s)
			if err != nil {
				return SignatureAlgorithm{}, err
			}
		case 1:
			sa.PaddingMode, err = ParsePaddingMode(s)
			if err != nil {
				return SignatureAlgorithm{}, err
			}
		case 2:
			sa.Hash, err = ParseHash(s)
			if err != nil {
				return SignatureAlgorithm{}, err
			}
		}
	}

	return sa, nil
}

// String 返回以连字符分隔的签名算法配置。
func (sa SignatureAlgorithm) String() string {
	return fmt.Sprintf("%s-%s-%s", sa.AsymmetricEncryption, sa.PaddingMode, sa.Hash)
}

// Read 将签名算法编码到 p。
func (sa SignatureAlgorithm) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint8(uint8(sa.AsymmetricEncryption)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(sa.PaddingMode)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteUint8(uint8(sa.Hash)); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码签名算法。
func (sa *SignatureAlgorithm) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	asymmetricEncryption, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	sa.AsymmetricEncryption = AsymmetricEncryption(asymmetricEncryption)

	paddingMode, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	sa.PaddingMode = PaddingMode(paddingMode)

	hash, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	sa.Hash = Hash(hash)

	return bs.BytesRead(), nil
}

// Size 返回签名算法的固定编码字节数。
func (SignatureAlgorithm) Size() int {
	return binaryutil.SizeofUint8 + binaryutil.SizeofUint8 + binaryutil.SizeofUint8
}

// MsgECDHESecretKeyExchange 携带 ECDHE 临时公钥、加密参数及可选签名。
// 直接通过 Write 或 Unmarshal 解码时，字节字段会引用输入切片；输入将被复用或修改时应先 Clone。
// Decoder.Decode 返回的消息不引用调用方输入。
type MsgECDHESecretKeyExchange struct {
	NamedCurve         NamedCurve         // ECDHE 命名曲线。
	PublicKey          []byte             // 临时公钥。
	IV                 []byte             // 分组密码初始化向量。
	Nonce              []byte             // 流密码或 AEAD 的初始 nonce。
	NonceStep          []byte             // 每包更新 nonce 使用的步进值。
	SignatureAlgorithm SignatureAlgorithm // 临时公钥签名算法。
	Signature          []byte             // 临时公钥及协商参数的签名。
}

// Read 将密钥交换消息编码到 p。
func (m MsgECDHESecretKeyExchange) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteUint8(uint8(m.NamedCurve)); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.PublicKey); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.IV); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Nonce); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.NonceStep); err != nil {
		return bs.BytesWritten(), err
	}
	if _, err := binaryutil.CopyToByteStream(&bs, m.SignatureAlgorithm); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteBytes(m.Signature); err != nil {
		return bs.BytesWritten(), err
	}
	return bs.BytesWritten(), io.EOF
}

// Write 从 p 解码密钥交换消息，字节字段会引用 p。
func (m *MsgECDHESecretKeyExchange) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	namedCurve, err := bs.ReadUint8()
	if err != nil {
		return bs.BytesRead(), err
	}
	m.NamedCurve = NamedCurve(namedCurve)

	m.PublicKey, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.IV, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.Nonce, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	m.NonceStep, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	if _, err := bs.WriteTo(&m.SignatureAlgorithm); err != nil {
		return bs.BytesRead(), err
	}

	m.Signature, err = bs.ReadBytesRef()
	if err != nil {
		return bs.BytesRead(), err
	}

	return bs.BytesRead(), nil
}

// Size 返回密钥交换消息编码后的字节数。
func (m MsgECDHESecretKeyExchange) Size() int {
	return binaryutil.SizeofUint8 + binaryutil.SizeofBytes(m.PublicKey) + binaryutil.SizeofBytes(m.IV) +
		binaryutil.SizeofBytes(m.Nonce) + binaryutil.SizeofBytes(m.NonceStep) + m.SignatureAlgorithm.Size() +
		binaryutil.SizeofBytes(m.Signature)
}

// MsgId 返回 ECDHE 密钥交换消息的内置类型 ID。
func (MsgECDHESecretKeyExchange) MsgId() MsgId {
	return MsgId_ECDHESecretKeyExchange
}

// Clone 深复制所有引用型字段。
func (m MsgECDHESecretKeyExchange) Clone() Msg {
	return &MsgECDHESecretKeyExchange{
		NamedCurve:         m.NamedCurve,
		PublicKey:          bytes.Clone(m.PublicKey),
		IV:                 bytes.Clone(m.IV),
		Nonce:              bytes.Clone(m.Nonce),
		NonceStep:          bytes.Clone(m.NonceStep),
		SignatureAlgorithm: m.SignatureAlgorithm,
		Signature:          bytes.Clone(m.Signature),
	}
}

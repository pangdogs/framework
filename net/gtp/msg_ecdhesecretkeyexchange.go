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

// ECDHESecretKeyExchange消息标志位
const (
	Flag_Signature Flag = 1 << (iota + Flag_Customize) // 有签名数据，在双方的ECDHE秘钥交换消息中携带，表示可以验证对方签名
)

// SignatureAlgorithm 签名算法
type SignatureAlgorithm struct {
	AsymmetricEncryption AsymmetricEncryption // 非对称加密算法
	PaddingMode          PaddingMode          // 填充方案
	Hash                 Hash                 // 摘要函数
}

// ParseSignatureAlgorithm 解析配置字串
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

// String implements fmt.Stringer
func (sa SignatureAlgorithm) String() string {
	return fmt.Sprintf("%s-%s-%s", sa.AsymmetricEncryption, sa.PaddingMode, sa.Hash)
}

// Read implements io.Reader
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

// Write implements io.Writer
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

// Size 大小
func (SignatureAlgorithm) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}

// MsgECDHESecretKeyExchange ECDHE秘钥交换消息，利用(g^a mod p)^b mod p == (g^b mod p)^a mod p等式，交换秘钥
// （注意：为了提高解码性能，减少内存碎片，解码string与bytes字段时均使用引用类型，引用字节池中的bytes，GC时会被归还字节池，不要直接持有此类型字段）
type MsgECDHESecretKeyExchange struct {
	NamedCurve         NamedCurve         // 曲线类型
	PublicKey          []byte             // 公钥
	IV                 []byte             // iv
	Nonce              []byte             // nonce
	NonceStep          []byte             // nonce step
	SignatureAlgorithm SignatureAlgorithm // 签名算法
	Signature          []byte             // 签名
}

// Read implements io.Reader
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

// Write implements io.Writer
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

// Size 大小
func (m MsgECDHESecretKeyExchange) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofBytes(m.PublicKey) + binaryutil.SizeofBytes(m.IV) +
		binaryutil.SizeofBytes(m.Nonce) + binaryutil.SizeofBytes(m.NonceStep) + m.SignatureAlgorithm.Size() +
		binaryutil.SizeofBytes(m.Signature)
}

// MsgId 消息Id
func (MsgECDHESecretKeyExchange) MsgId() MsgId {
	return MsgId_ECDHESecretKeyExchange
}

// Clone 克隆消息对象
func (m MsgECDHESecretKeyExchange) Clone() MsgReader {
	return MsgECDHESecretKeyExchange{
		NamedCurve:         m.NamedCurve,
		PublicKey:          bytes.Clone(m.PublicKey),
		IV:                 bytes.Clone(m.IV),
		Nonce:              bytes.Clone(m.Nonce),
		NonceStep:          bytes.Clone(m.NonceStep),
		SignatureAlgorithm: m.SignatureAlgorithm,
		Signature:          bytes.Clone(m.Signature),
	}
}

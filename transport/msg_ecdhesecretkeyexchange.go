package transport

import (
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// ECDHESecretKeyExchange消息标志位
const (
	Flag_Signature Flag = 1 << (iota + Flag_Customize) // 有签名数据
)

// SignatureAlgorithm 签名算法
type SignatureAlgorithm struct {
	AsymmetricEncryption AsymmetricEncryption // 非对称加密算法
	PaddingMode          PaddingMode          // 填充方案
	Hash                 Hash                 // 摘要函数
}

func (sa *SignatureAlgorithm) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint8(uint8(sa.AsymmetricEncryption)); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(uint8(sa.PaddingMode)); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(uint8(sa.Hash)); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (sa *SignatureAlgorithm) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	asymmetricEncryption, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	paddingMode, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	hash, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	sa.AsymmetricEncryption = AsymmetricEncryption(asymmetricEncryption)
	sa.PaddingMode = PaddingMode(paddingMode)
	sa.Hash = Hash(hash)
	return bs.BytesRead(), nil
}

func (sa *SignatureAlgorithm) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}

// MsgECDHESecretKeyExchange ECDHE秘钥交换消息，利用(g^a mod p)^b mod p == (g^b mod p)^a mod p等式，交换秘钥
type MsgECDHESecretKeyExchange struct {
	NamedCurve         NamedCurve         // 曲线类型
	PublicKey          []byte             // 公钥
	IV                 []byte             // iv
	Nonce              []byte             // nonce
	SignatureAlgorithm SignatureAlgorithm // 签名算法
	Signature          []byte             // 签名
}

func (m *MsgECDHESecretKeyExchange) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint8(uint8(m.NamedCurve)); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.PublicKey); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.IV); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Nonce); err != nil {
		return 0, err
	}
	if _, err := bs.ReadFrom(&m.SignatureAlgorithm); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Signature); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgECDHESecretKeyExchange) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	namedCurve, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	publicKey, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	iv, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	nonce, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	signatureAlgorithm := SignatureAlgorithm{}
	if _, err := bs.WriteTo(&signatureAlgorithm); err != nil {
		return 0, err
	}
	signature, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.NamedCurve = NamedCurve(namedCurve)
	m.PublicKey = publicKey
	m.IV = iv
	m.Nonce = nonce
	m.SignatureAlgorithm = signatureAlgorithm
	m.Signature = signature
	return bs.BytesRead(), nil
}

func (m *MsgECDHESecretKeyExchange) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofBytes(m.PublicKey) +
		m.SignatureAlgorithm.Size() + binaryutil.SizeofBytes(m.Signature)
}

func (MsgECDHESecretKeyExchange) MsgId() MsgId {
	return MsgId_ECDHESecretKeyExchange
}

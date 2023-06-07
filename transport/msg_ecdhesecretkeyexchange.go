package transport

import (
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// SignatureAlgorithm 签名算法
type SignatureAlgorithm struct {
	AsymmetricEncryptMethod AsymmetricEncryptMethod // 非对称加密函数
	PaddingMode             PaddingMode             // 非对称加密算法填充方案
	HashMethod              HashMethod              // 摘要函数
}

func (sa *SignatureAlgorithm) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint8(sa.AsymmetricEncryptMethod); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(sa.PaddingMode); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(sa.HashMethod); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (sa *SignatureAlgorithm) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	asymmetricEncryptMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	paddingMode, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	hashMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	sa.AsymmetricEncryptMethod = asymmetricEncryptMethod
	sa.PaddingMode = paddingMode
	sa.HashMethod = hashMethod
	return bs.BytesRead(), nil
}

func (sa *SignatureAlgorithm) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}

// MsgECDHESecretKeyExchange ECDHE秘钥交换消息，利用(g^a mod p)^b mod p == (g^b mod p)^a mod p等式，交换秘钥
type MsgECDHESecretKeyExchange struct {
	NamedCurve         NamedCurve         // 曲线类型
	PublicKey          []byte             // 公钥
	SignatureAlgorithm SignatureAlgorithm // 签名算法
	Signature          []byte             // 签名
}

func (m *MsgECDHESecretKeyExchange) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint8(m.NamedCurve); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.PublicKey); err != nil {
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
	signatureAlgorithm := SignatureAlgorithm{}
	if _, err := bs.WriteTo(&signatureAlgorithm); err != nil {
		return 0, err
	}
	signature, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.NamedCurve = namedCurve
	m.PublicKey = publicKey
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

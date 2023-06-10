package transport

import (
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// Hello消息标志位
const (
	Flag_HelloDone     Flag = 1 << (iota + Flag_Customize) // Hello完成，在服务端返回的Hello消息中携带，表示初步认可客户端连接
	Flag_EnableEncrypt                                     // 开启加密（协议优先考虑性能，要求安全性请直接使用TLS加密链路），在服务端返回的Hello消息中携带，表示链路需要加密，需要执行秘钥交换流程
	Flag_EnableAuth                                        // 开启鉴权（基于token鉴权），在服务端返回的Hello消息中携带，表示链路需要认证，需要执行鉴权流程
)

// CipherSuite 密码学套件
type CipherSuite struct {
	SecretKeyExchangeMethod SecretKeyExchangeMethod // 秘钥交换函数
	SymmetricEncryptMethod  SymmetricEncryptMethod  // 对称加密函数
	BlockCipherMode         BlockCipherMode         // 对称加密算法分组模式
	HashMethod              HashMethod              // 摘要函数
}

func (cs *CipherSuite) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint8(cs.SecretKeyExchangeMethod); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(cs.SymmetricEncryptMethod); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(cs.BlockCipherMode); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(cs.HashMethod); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (cs *CipherSuite) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	secretKeyExchangeMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	symmetricEncryptMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	blockCipherMode, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	hashMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	cs.SecretKeyExchangeMethod = secretKeyExchangeMethod
	cs.SymmetricEncryptMethod = symmetricEncryptMethod
	cs.BlockCipherMode = blockCipherMode
	cs.HashMethod = hashMethod
	return bs.BytesRead(), nil
}

func (cs *CipherSuite) Size() int {
	return binaryutil.SizeofUint8() + binaryutil.SizeofUint8() +
		binaryutil.SizeofUint8() + binaryutil.SizeofUint8()
}

// MsgHello Hello消息
type MsgHello struct {
	Version           Version           // 协议版本
	SessionId         []byte            // 会话Id，如果客户端上传空值，服务端将会分配新会话，如果非空值，服务端将尝试查找会话，查找失败会重置链路
	Random            []byte            // 随机数，用于秘钥交换
	CipherSuite       CipherSuite       // 密码学套件，客户端提交的密码学套件建议，服务端可能不采纳，以服务端返回的为准，若客户端不支持，直接切断链路
	CompressionMethod CompressionMethod // 压缩函数，客户端提交的压缩函数建议，服务端可能不采纳，以服务端返回的为准，若客户端不支持，直接切断链路
	Extensions        []byte            // 扩展内容
}

func (m *MsgHello) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	if err := bs.WriteUint16(m.Version); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.SessionId); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Random); err != nil {
		return 0, err
	}
	if _, err := bs.ReadFrom(&m.CipherSuite); err != nil {
		return 0, err
	}
	if err := bs.WriteUint8(m.CompressionMethod); err != nil {
		return 0, err
	}
	if err := bs.WriteBytes(m.Extensions); err != nil {
		return 0, err
	}
	return bs.BytesWritten(), nil
}

func (m *MsgHello) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)
	version, err := bs.ReadUint16()
	if err != nil {
		return 0, err
	}
	sessionId, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	randomn, err := bs.ReadBytes()
	if err != nil {
		return 0, err
	}
	cipherSuite := CipherSuite{}
	if _, err := bs.WriteTo(&cipherSuite); err != nil {
		return 0, err
	}
	compressionMethod, err := bs.ReadUint8()
	if err != nil {
		return 0, err
	}
	extensions, err := bs.ReadBytesRef()
	if err != nil {
		return 0, err
	}
	m.Version = version
	m.SessionId = sessionId
	m.Random = randomn
	m.CipherSuite = cipherSuite
	m.CompressionMethod = compressionMethod
	m.Extensions = extensions
	return bs.BytesRead(), nil
}

func (m *MsgHello) Size() int {
	return binaryutil.SizeofUint16() + binaryutil.SizeofBytes(m.SessionId) + binaryutil.SizeofBytes(m.Random) +
		m.CipherSuite.Size() + binaryutil.SizeofUint8() + binaryutil.SizeofBytes(m.Extensions)
}

func (MsgHello) MsgId() MsgId {
	return MsgId_Hello
}

package tcp

import (
	"crypto"
	"crypto/tls"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"math/big"
	"net"
	"time"
)

type (
	Auth = func(token string, extensions []byte) error
)

type WithOption struct{}

type GateOptions struct {
	Endpoints                []string                     // 所有监听地址
	TLSConfig                *tls.Config                  // TLS配置，nil表示不使用TLS加密链路
	TCPNoDelay               *bool                        // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck              *bool                        // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf               *int                         // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf               *int                         // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                *int                         // TCP的PLinger选项，nil表示使用系统默认值
	Timeout                  time.Duration                // 网络io超时时间
	RetryTimes               int                          // 网络io超时后的重试次数
	SequencedBuffCap         int                          // 时序缓存容量（字节）
	MsgCreator               codec.IMsgCreator            // 消息构建器
	AgreeClientProposal      bool                         // 是否同意使用客户端建议的加密与压缩方案
	CipherSuite              transport.CipherSuite        // 密码学套件
	NonceStep                *big.Int                     // 使用需要nonce的加密算法时，每次加解密自增值
	ECDHENamedCurve          transport.NamedCurve         // ECDHE交换秘钥时使用的曲线类型
	SignatureAlgorithm       transport.SignatureAlgorithm // 签名算法
	SignaturePrivateKey      crypto.PrivateKey            // 签名用的私钥
	VerifyClientSignature    bool                         // 验证客户端签名
	VerifySignaturePublicKey crypto.PublicKey             // 验证客户端签名用的公钥
	Compression              transport.Compression        // 压缩函数
	CompressedSize           int                          // 启用压缩阀值（字节），<=0表示不开启
	Auth                     Auth                         // 鉴权函数
}

type GateOption func(options *GateOptions)

func (WithOption) Default() GateOption {
	return func(options *GateOptions) {
		WithOption{}.Endpoints("0.0.0.0:0")(options)
		WithOption{}.TLSConfig(nil)(options)
		WithOption{}.TCPNoDelay(nil)(options)
		WithOption{}.TCPQuickAck(nil)(options)
		WithOption{}.TCPRecvBuf(nil)(options)
		WithOption{}.TCPSendBuf(nil)(options)
		WithOption{}.TCPLinger(nil)(options)
		WithOption{}.Timeout(3 * time.Second)(options)
		WithOption{}.RetryTimes(3)(options)
		WithOption{}.SequencedBuffCap(1024 * 128)(options)
		WithOption{}.MsgCreator(codec.DefaultMsgCreator())(options)
		WithOption{}.AgreeClientProposal(false)(options)
		WithOption{}.CipherSuite(transport.CipherSuite{
			SecretKeyExchange:   transport.SecretKeyExchange_ECDHE,
			SymmetricEncryption: transport.SymmetricEncryption_AES,
			BlockCipherMode:     transport.BlockCipherMode_CTR,
			PaddingMode:         transport.PaddingMode_None,
			MACHash:             transport.Hash_Fnv1a32,
		})(options)
		WithOption{}.NonceStep(new(big.Int).SetInt64(1))
		WithOption{}.ECDHENamedCurve(transport.NamedCurve_X25519)(options)
		WithOption{}.SignatureAlgorithm(transport.SignatureAlgorithm{
			AsymmetricEncryption: transport.AsymmetricEncryption_None,
			PaddingMode:          transport.PaddingMode_None,
			Hash:                 transport.Hash_None,
		})(options)
		WithOption{}.SignaturePrivateKey(nil)
		WithOption{}.VerifyClientSignature(false)
		WithOption{}.VerifySignaturePublicKey(nil)
		WithOption{}.Compression(transport.Compression_Brotli)(options)
		WithOption{}.CompressedSize(1024 * 32)(options)
		WithOption{}.Auth(nil)(options)
	}
}

func (WithOption) Endpoints(endpoints ...string) GateOption {
	return func(options *GateOptions) {
		for _, endpoint := range endpoints {
			if _, _, err := net.SplitHostPort(endpoint); err != nil {
				panic(err)
			}
		}
		options.Endpoints = endpoints
	}
}

func (WithOption) TLSConfig(tlsConfig *tls.Config) GateOption {
	return func(options *GateOptions) {
		options.TLSConfig = tlsConfig
	}
}

func (WithOption) TCPNoDelay(b *bool) GateOption {
	return func(options *GateOptions) {
		options.TCPNoDelay = b
	}
}

func (WithOption) TCPQuickAck(b *bool) GateOption {
	return func(options *GateOptions) {
		options.TCPQuickAck = b
	}
}

func (WithOption) TCPRecvBuf(size *int) GateOption {
	return func(options *GateOptions) {
		options.TCPRecvBuf = size
	}
}

func (WithOption) TCPSendBuf(size *int) GateOption {
	return func(options *GateOptions) {
		options.TCPSendBuf = size
	}
}

func (WithOption) TCPLinger(sec *int) GateOption {
	return func(options *GateOptions) {
		options.TCPLinger = sec
	}
}

func (WithOption) Timeout(d time.Duration) GateOption {
	return func(options *GateOptions) {
		options.Timeout = d
	}
}

func (WithOption) RetryTimes(times int) GateOption {
	return func(options *GateOptions) {
		options.RetryTimes = times
	}
}

func (WithOption) SequencedBuffCap(cap int) GateOption {
	return func(options *GateOptions) {
		options.SequencedBuffCap = cap
	}
}

func (WithOption) MsgCreator(mc codec.IMsgCreator) GateOption {
	return func(options *GateOptions) {
		if mc == nil {
			panic("option MsgCreator can't be assigned to nil")
		}
		options.MsgCreator = mc
	}
}

func (WithOption) AgreeClientProposal(b bool) GateOption {
	return func(options *GateOptions) {
		options.AgreeClientProposal = b
	}
}

func (WithOption) CipherSuite(cs transport.CipherSuite) GateOption {
	return func(options *GateOptions) {
		options.CipherSuite = cs
	}
}

func (WithOption) NonceStep(v *big.Int) GateOption {
	return func(options *GateOptions) {
		options.NonceStep = v
	}
}

func (WithOption) ECDHENamedCurve(nc transport.NamedCurve) GateOption {
	return func(options *GateOptions) {
		options.ECDHENamedCurve = nc
	}
}

func (WithOption) SignatureAlgorithm(sa transport.SignatureAlgorithm) GateOption {
	return func(options *GateOptions) {
		options.SignatureAlgorithm = sa
	}
}

func (WithOption) SignaturePrivateKey(priv crypto.PrivateKey) GateOption {
	return func(options *GateOptions) {
		options.SignaturePrivateKey = priv
	}
}

func (WithOption) VerifyClientSignature(b bool) GateOption {
	return func(options *GateOptions) {
		options.VerifyClientSignature = b
	}
}

func (WithOption) VerifySignaturePublicKey(pub crypto.PublicKey) GateOption {
	return func(options *GateOptions) {
		options.VerifySignaturePublicKey = pub
	}
}

func (WithOption) Compression(c transport.Compression) GateOption {
	return func(options *GateOptions) {
		options.Compression = c
	}
}

func (WithOption) CompressedSize(size int) GateOption {
	return func(options *GateOptions) {
		options.CompressedSize = size
	}
}

func (WithOption) Auth(fn Auth) GateOption {
	return func(options *GateOptions) {
		options.Auth = fn
	}
}

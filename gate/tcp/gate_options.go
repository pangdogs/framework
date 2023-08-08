package tcp

import (
	"crypto"
	"crypto/tls"
	"kit.golaxy.org/plugins/gate"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"math/big"
	"net"
	"time"
)

type (
	ClientAuthHandler          = func(conn net.Conn, token string, extensions []byte) error // 客户端鉴权鉴权处理器
	SessionStateChangedHandler = gate.StateChangedHandler                                   // 会话状态变化的处理器
	SessionRecvDataHandler     = gate.RecvDataHandler                                       // 会话接收的数据的处理器
	SessionRecvEventHandler    = gate.RecvEventHandler                                      // 会话接收的自定义事件的处理器
)

type WithOption struct{}

type GateOptions struct {
	Endpoints                      []string                     // 所有监听地址
	TLSConfig                      *tls.Config                  // TLS配置，nil表示不使用TLS加密链路
	TCPNoDelay                     *bool                        // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                    *bool                        // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                     *int                         // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                     *int                         // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                      *int                         // TCP的PLinger选项，nil表示使用系统默认值
	IOTimeout                      time.Duration                // 网络io超时时间
	IORetryTimes                   int                          // 网络io超时后的重试次数
	IOSequencedBuffCap             int                          // 网络io时序缓存容量（字节）
	DecoderMsgCreator              codec.IMsgCreator            // 消息包解码器的消息构建器
	AgreeClientEncryptionProposal  bool                         // 是否同意使用客户端建议的加密方案
	EncCipherSuite                 transport.CipherSuite        // 加密通信中的密码学套件
	EncNonceStep                   *big.Int                     // 加密通信中，使用需要nonce的加密算法时，每次加解密自增值
	EncECDHENamedCurve             transport.NamedCurve         // 加密通信中，在ECDHE交换秘钥时使用的曲线类型
	EncSignatureAlgorithm          transport.SignatureAlgorithm // 加密通信中的签名算法
	EncSignaturePrivateKey         crypto.PrivateKey            // 加密通信中，签名用的私钥
	EncVerifyClientSignature       bool                         // 加密通信中，是否验证客户端签名
	EncVerifySignaturePublicKey    crypto.PublicKey             // 加密通信中，验证客户端签名用的公钥
	AgreeClientCompressionProposal bool                         // 是否同意使用客户端建议的压缩方案
	Compression                    transport.Compression        // 通信中的压缩函数
	CompressedSize                 int                          // 通信中启用压缩阀值（字节），<=0表示不开启
	SessionInactiveTimeout         time.Duration                // 会话不活跃后的超时时间
	ClientAuthHandlers             []ClientAuthHandler          // 客户端鉴权鉴权处理器列表
	SessionStateChangedHandlers    []SessionStateChangedHandler // 会话状态变化的处理器列表（优先级高于会话的处理器）
	SessionRecvDataHandlers        []SessionRecvDataHandler     // 会话接收的数据的处理器列表（优先级高于会话的处理器）
	SessionRecvEventHandlers       []SessionRecvEventHandler    // 会话接收的自定义事件的处理器列表（优先级高于会话的处理器）
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
		WithOption{}.IOTimeout(3 * time.Second)(options)
		WithOption{}.IORetryTimes(3)(options)
		WithOption{}.IOSequencedBuffCap(1024 * 128)(options)
		WithOption{}.DecoderMsgCreator(codec.DefaultMsgCreator())(options)
		WithOption{}.AgreeClientEncryptionProposal(false)(options)
		WithOption{}.EncCipherSuite(transport.CipherSuite{
			SecretKeyExchange:   transport.SecretKeyExchange_ECDHE,
			SymmetricEncryption: transport.SymmetricEncryption_AES,
			BlockCipherMode:     transport.BlockCipherMode_CTR,
			PaddingMode:         transport.PaddingMode_None,
			MACHash:             transport.Hash_Fnv1a32,
		})(options)
		WithOption{}.EncNonceStep(new(big.Int).SetInt64(1))
		WithOption{}.EncECDHENamedCurve(transport.NamedCurve_X25519)(options)
		WithOption{}.EncSignatureAlgorithm(transport.SignatureAlgorithm{
			AsymmetricEncryption: transport.AsymmetricEncryption_None,
			PaddingMode:          transport.PaddingMode_None,
			Hash:                 transport.Hash_None,
		})(options)
		WithOption{}.EncSignaturePrivateKey(nil)
		WithOption{}.EncVerifyClientSignature(false)
		WithOption{}.EncVerifySignaturePublicKey(nil)
		WithOption{}.AgreeClientCompressionProposal(false)
		WithOption{}.Compression(transport.Compression_Brotli)(options)
		WithOption{}.CompressedSize(1024 * 32)(options)
		WithOption{}.SessionInactiveTimeout(60 * time.Second)(options)
		WithOption{}.ClientAuthHandlers(nil)(options)
		WithOption{}.SessionStateChangedHandlers(nil)(options)
		WithOption{}.SessionRecvDataHandlers(nil)(options)
		WithOption{}.SessionRecvEventHandlers(nil)(options)
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

func (WithOption) IOTimeout(d time.Duration) GateOption {
	return func(options *GateOptions) {
		options.IOTimeout = d
	}
}

func (WithOption) IORetryTimes(times int) GateOption {
	return func(options *GateOptions) {
		options.IORetryTimes = times
	}
}

func (WithOption) IOSequencedBuffCap(cap int) GateOption {
	return func(options *GateOptions) {
		options.IOSequencedBuffCap = cap
	}
}

func (WithOption) DecoderMsgCreator(mc codec.IMsgCreator) GateOption {
	return func(options *GateOptions) {
		if mc == nil {
			panic("option DecoderMsgCreator can't be assigned to nil")
		}
		options.DecoderMsgCreator = mc
	}
}

func (WithOption) AgreeClientEncryptionProposal(b bool) GateOption {
	return func(options *GateOptions) {
		options.AgreeClientEncryptionProposal = b
	}
}

func (WithOption) EncCipherSuite(cs transport.CipherSuite) GateOption {
	return func(options *GateOptions) {
		options.EncCipherSuite = cs
	}
}

func (WithOption) EncNonceStep(v *big.Int) GateOption {
	return func(options *GateOptions) {
		options.EncNonceStep = v
	}
}

func (WithOption) EncECDHENamedCurve(nc transport.NamedCurve) GateOption {
	return func(options *GateOptions) {
		options.EncECDHENamedCurve = nc
	}
}

func (WithOption) EncSignatureAlgorithm(sa transport.SignatureAlgorithm) GateOption {
	return func(options *GateOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

func (WithOption) EncSignaturePrivateKey(priv crypto.PrivateKey) GateOption {
	return func(options *GateOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

func (WithOption) EncVerifyClientSignature(b bool) GateOption {
	return func(options *GateOptions) {
		options.EncVerifyClientSignature = b
	}
}

func (WithOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) GateOption {
	return func(options *GateOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

func (WithOption) AgreeClientCompressionProposal(b bool) GateOption {
	return func(options *GateOptions) {
		options.AgreeClientCompressionProposal = b
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

func (WithOption) SessionInactiveTimeout(d time.Duration) GateOption {
	return func(options *GateOptions) {
		options.SessionInactiveTimeout = d
	}
}

func (WithOption) ClientAuthHandlers(handlers []ClientAuthHandler) GateOption {
	return func(options *GateOptions) {
		options.ClientAuthHandlers = handlers
	}
}

func (WithOption) SessionStateChangedHandlers(handlers []SessionStateChangedHandler) GateOption {
	return func(options *GateOptions) {
		options.SessionStateChangedHandlers = handlers
	}
}

func (WithOption) SessionRecvDataHandlers(handlers []SessionRecvDataHandler) GateOption {
	return func(options *GateOptions) {
		options.SessionRecvDataHandlers = handlers
	}
}

func (WithOption) SessionRecvEventHandlers(handlers ...SessionRecvEventHandler) GateOption {
	return func(options *GateOptions) {
		options.SessionRecvEventHandlers = handlers
	}
}

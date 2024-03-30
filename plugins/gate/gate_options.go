package gate

import (
	"crypto"
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"math/big"
	"net"
	"time"
)

type (
	AuthClientHandler          = generic.DelegateFunc5[IGate, net.Conn, string, string, []byte, error] // 客户端鉴权处理器（args: [gate, conn, userId, token, extensions], ret: [error]）
	SessionStateChangedHandler = generic.DelegateAction3[ISession, SessionState, SessionState]         // 会话状态变化的处理器（args: [session, curState, lastState]）
	SessionRecvDataHandler     = generic.DelegateFunc2[ISession, []byte, error]                        // 会话接收的数据的处理器
	SessionRecvEventHandler    = generic.DelegateFunc2[ISession, transport.Event[gtp.Msg], error]      // 会话接收的自定义事件的处理器
)

type GateOptions struct {
	TCPAddress                     string                     // TCP监听地址
	TCPNoDelay                     *bool                      // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                    *bool                      // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                     *int                       // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                     *int                       // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                      *int                       // TCP的PLinger选项，nil表示使用系统默认值
	TCPTLSConfig                   *tls.Config                // TCP的TLS配置，nil表示不使用TLS加密链路
	WebSocketAddress               string                     // WebSocket监听地址
	WebSocketTLSConfig             *tls.Config                // TCP的TLS配置，nil表示不使用TLS加密链路
	IOTimeout                      time.Duration              // 网络io超时时间
	IORetryTimes                   int                        // 网络io超时后的重试次数
	IOBufferCap                    int                        // 网络io缓存容量（字节）
	DecoderMsgCreator              gtp.IMsgCreator            // 消息包解码器的消息构建器
	AgreeClientEncryptionProposal  bool                       // 是否同意使用客户端建议的加密方案
	EncCipherSuite                 gtp.CipherSuite            // 加密通信中的密码学套件
	EncNonceStep                   *big.Int                   // 加密通信中，使用需要nonce的加密算法时，每次加解密自增值
	EncECDHENamedCurve             gtp.NamedCurve             // 加密通信中，在ECDHE交换秘钥时使用的曲线类型
	EncSignatureAlgorithm          gtp.SignatureAlgorithm     // 加密通信中的签名算法
	EncSignaturePrivateKey         crypto.PrivateKey          // 加密通信中，签名用的私钥
	EncVerifyClientSignature       bool                       // 加密通信中，是否验证客户端签名
	EncVerifySignaturePublicKey    crypto.PublicKey           // 加密通信中，验证客户端签名用的公钥
	AgreeClientCompressionProposal bool                       // 是否同意使用客户端建议的压缩方案
	Compression                    gtp.Compression            // 通信中的压缩函数
	CompressedSize                 int                        // 通信中启用压缩阀值（字节），<=0表示不开启
	AcceptTimeout                  time.Duration              // 接受连接超时时间
	AuthClientHandler              AuthClientHandler          // 客户端鉴权处理器
	SessionInactiveTimeout         time.Duration              // 会话不活跃后的超时时间
	SessionStateChangedHandler     SessionStateChangedHandler // 会话状态变化的处理器（优先级低于会话的处理器）
	SessionSendDataChanSize        int                        // 会话默认发送数据的channel的大小，<=0表示不使用channel
	SessionRecvDataChanSize        int                        // 会话默认接收数据的channel的大小，<=0表示不使用channel
	SessionRecvDataChanRecyclable  bool                       // 会话默认接收数据的channel中是否使用可回收字节对象
	SessionSendEventChanSize       int                        // 会话默认发送自定义事件的channel的大小，<=0表示不使用channel
	SessionRecvEventChanSize       int                        // 会话默认接收自定义事件的channel的大小，<=0表示不使用channel
	SessionRecvDataHandler         SessionRecvDataHandler     // 会话接收的数据的处理器（优先级低于会话的处理器）
	SessionRecvEventHandler        SessionRecvEventHandler    // 会话接收的自定义事件的处理器（优先级低于会话的处理器）
}

var With _GateOption

type _GateOption struct{}

func (_GateOption) Default() option.Setting[GateOptions] {
	return func(options *GateOptions) {
		With.TCPAddress("0.0.0.0:0")(options)
		With.TCPNoDelay(nil)(options)
		With.TCPQuickAck(nil)(options)
		With.TCPRecvBuf(nil)(options)
		With.TCPSendBuf(nil)(options)
		With.TCPLinger(nil)(options)
		With.TCPTLSConfig(nil)(options)
		With.WebSocketAddress("")(options)
		With.WebSocketTLSConfig(nil)(options)
		With.IOTimeout(3 * time.Second)(options)
		With.IORetryTimes(3)(options)
		With.IOBufferCap(1024 * 128)(options)
		With.DecoderMsgCreator(gtp.DefaultMsgCreator())(options)
		With.AgreeClientEncryptionProposal(false)(options)
		With.EncCipherSuite(gtp.CipherSuite{
			SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
			SymmetricEncryption: gtp.SymmetricEncryption_AES,
			BlockCipherMode:     gtp.BlockCipherMode_CTR,
			PaddingMode:         gtp.PaddingMode_None,
			MACHash:             gtp.Hash_Fnv1a32,
		})(options)
		With.EncNonceStep(big.NewInt(1))(options)
		With.EncECDHENamedCurve(gtp.NamedCurve_X25519)(options)
		With.EncSignatureAlgorithm(gtp.SignatureAlgorithm{
			AsymmetricEncryption: gtp.AsymmetricEncryption_None,
			PaddingMode:          gtp.PaddingMode_None,
			Hash:                 gtp.Hash_None,
		})(options)
		With.EncSignaturePrivateKey(nil)(options)
		With.EncVerifyClientSignature(false)(options)
		With.EncVerifySignaturePublicKey(nil)(options)
		With.AgreeClientCompressionProposal(false)(options)
		With.Compression(gtp.Compression_Brotli)(options)
		With.CompressedSize(1024 * 32)(options)
		With.AcceptTimeout(5 * time.Second)(options)
		With.AuthClientHandler(nil)(options)
		With.SessionInactiveTimeout(time.Minute)(options)
		With.SessionStateChangedHandler(nil)(options)
		With.SessionSendDataChanSize(0)(options)
		With.SessionRecvDataChanSize(0, false)(options)
		With.SessionSendEventChanSize(0)(options)
		With.SessionRecvEventChanSize(0)(options)
		With.SessionRecvDataHandler(nil)(options)
		With.SessionRecvEventHandler(nil)(options)
	}
}

func (_GateOption) TCPAddress(addr string) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if addr != "" {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.TCPAddress = addr
	}
}

func (_GateOption) TCPNoDelay(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPNoDelay = b
	}
}

func (_GateOption) TCPQuickAck(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPQuickAck = b
	}
}

func (_GateOption) TCPRecvBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPRecvBuf = size
	}
}

func (_GateOption) TCPSendBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPSendBuf = size
	}
}

func (_GateOption) TCPLinger(sec *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPLinger = sec
	}
}

func (_GateOption) TCPTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPTLSConfig = tlsConfig
	}
}

func (_GateOption) WebSocketAddress(addr string) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if addr != "" {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				panic(fmt.Errorf("%w: %w", core.ErrArgs, err))
			}
		}
		options.WebSocketAddress = addr
	}
}

func (_GateOption) WebSocketTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.WebSocketTLSConfig = tlsConfig
	}
}

func (_GateOption) IOTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.IOTimeout = d
	}
}

func (_GateOption) IORetryTimes(times int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.IORetryTimes = times
	}
}

func (_GateOption) IOBufferCap(cap int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.IOBufferCap = cap
	}
}

func (_GateOption) DecoderMsgCreator(mc gtp.IMsgCreator) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if mc == nil {
			panic(fmt.Errorf("%w: option DecoderMsgCreator can't be assigned to nil", core.ErrArgs))
		}
		options.DecoderMsgCreator = mc
	}
}

func (_GateOption) AgreeClientEncryptionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientEncryptionProposal = b
	}
}

func (_GateOption) EncCipherSuite(cs gtp.CipherSuite) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncCipherSuite = cs
	}
}

func (_GateOption) EncNonceStep(v *big.Int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncNonceStep = v
	}
}

func (_GateOption) EncECDHENamedCurve(nc gtp.NamedCurve) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncECDHENamedCurve = nc
	}
}

func (_GateOption) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

func (_GateOption) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

func (_GateOption) EncVerifyClientSignature(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifyClientSignature = b
	}
}

func (_GateOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

func (_GateOption) AgreeClientCompressionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientCompressionProposal = b
	}
}

func (_GateOption) Compression(c gtp.Compression) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.Compression = c
	}
}

func (_GateOption) CompressedSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.CompressedSize = size
	}
}

func (_GateOption) AcceptTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AcceptTimeout = d
	}
}

func (_GateOption) AuthClientHandler(handler AuthClientHandler) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AuthClientHandler = handler
	}
}

func (_GateOption) SessionInactiveTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionInactiveTimeout = d
	}
}

func (_GateOption) SessionStateChangedHandler(handler SessionStateChangedHandler) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionStateChangedHandler = handler
	}
}

func (_GateOption) SessionSendDataChanSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionSendDataChanSize = size
	}
}

func (_GateOption) SessionRecvDataChanSize(size int, recyclable bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionRecvDataChanSize = size
		options.SessionRecvDataChanRecyclable = recyclable
	}
}

func (_GateOption) SessionSendEventChanSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionSendEventChanSize = size
	}
}

func (_GateOption) SessionRecvEventChanSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionRecvEventChanSize = size
	}
}

func (_GateOption) SessionRecvDataHandler(handler SessionRecvDataHandler) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionRecvDataHandler = handler
	}
}

func (_GateOption) SessionRecvEventHandler(handler SessionRecvEventHandler) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.SessionRecvEventHandler = handler
	}
}

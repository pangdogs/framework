package cli

import (
	"crypto"
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/core/util/generic"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/util/binaryutil"
	"go.uber.org/zap"
	"time"
)

type (
	RecvDataHandler  = generic.DelegateFunc1[[]byte, error]
	RecvEventHandler = transport.EventHandler
)

type NetProtocol int32

const (
	TCP NetProtocol = iota
	WebSocket
)

type ClientOptions struct {
	NetProtocol                 NetProtocol                         // 使用的网络协议（TCP/WebSocket）
	TCPNoDelay                  *bool                               // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                 *bool                               // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                  *int                                // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                  *int                                // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                   *int                                // TCP的PLinger选项，nil表示使用系统默认值
	WebSocketOrigin             string                              // WebSocket的Origin地址，不填将会自动生成
	TLSConfig                   *tls.Config                         // TLS配置，nil表示不使用TLS加密链路
	IOTimeout                   time.Duration                       // 网络io超时时间
	IORetryTimes                int                                 // 网络io超时后的重试次数
	IOBufferCap                 int                                 // 网络io缓存容量（字节）
	DecoderMsgCreator           gtp.IMsgCreator                     // 消息包解码器的消息构建器
	EncCipherSuite              gtp.CipherSuite                     // 加密通信中的密码学套件
	EncSignatureAlgorithm       gtp.SignatureAlgorithm              // 加密通信中的签名算法
	EncSignaturePrivateKey      crypto.PrivateKey                   // 加密通信中，签名用的私钥
	EncVerifyServerSignature    bool                                // 加密通信中，是否验证服务端签名
	EncVerifySignaturePublicKey crypto.PublicKey                    // 加密通信中，验证服务端签名用的公钥
	Compression                 gtp.Compression                     // 通信中的压缩函数
	CompressedSize              int                                 // 通信中启用压缩阀值（字节），<=0表示不开启
	AutoReconnect               bool                                // 开启自动重连
	AutoReconnectInterval       time.Duration                       // 自动重连的时间间隔
	AutoReconnectRetryTimes     int                                 // 自动重连的重试次数，<=0表示无限重试
	InactiveTimeout             time.Duration                       // 连接不活跃后的超时时间，开启自动重连后无效
	SendDataChan                chan binaryutil.RecycleBytes        // 发送数据的channel
	RecvDataChan                chan binaryutil.RecycleBytes        // 接收数据的channel
	RecvDataChanRecyclable      bool                                // 接收数据的channel中是否使用可回收字节对象
	SendEventChan               chan transport.Event[gtp.MsgReader] // 发送自定义事件的channel
	RecvEventChan               chan transport.Event[gtp.Msg]       // 接收自定义事件的channel
	RecvDataHandler             RecvDataHandler                     // 接收的数据的处理器
	RecvEventHandler            RecvEventHandler                    // 接收的自定义事件的处理器
	FutureTimeout               time.Duration                       // 异步模型Future超时时间
	AuthUserId                  string                              // 鉴权userid
	AuthToken                   string                              // 鉴权token
	AuthExtensions              []byte                              // 鉴权extensions
	ZapLogger                   *zap.Logger                         // zap日志
}

var With _Option

type _Option struct{}

func (_Option) Default() option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		With.NetProtocol(TCP)(options)
		With.TCPNoDelay(nil)(options)
		With.TCPQuickAck(nil)(options)
		With.TCPRecvBuf(nil)(options)
		With.TCPSendBuf(nil)(options)
		With.TCPLinger(nil)(options)
		With.WebSocketOrigin("")(options)
		With.TLSConfig(nil)(options)
		With.IOTimeout(3 * time.Second)(options)
		With.IORetryTimes(3)(options)
		With.IOBufferCap(1024 * 128)(options)
		With.DecoderMsgCreator(gtp.DefaultMsgCreator())(options)
		With.EncCipherSuite(gtp.CipherSuite{
			SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
			SymmetricEncryption: gtp.SymmetricEncryption_AES,
			BlockCipherMode:     gtp.BlockCipherMode_CTR,
			PaddingMode:         gtp.PaddingMode_None,
			MACHash:             gtp.Hash_Fnv1a32,
		})(options)
		With.EncSignatureAlgorithm(gtp.SignatureAlgorithm{
			AsymmetricEncryption: gtp.AsymmetricEncryption_None,
			PaddingMode:          gtp.PaddingMode_None,
			Hash:                 gtp.Hash_None,
		})(options)
		With.EncSignaturePrivateKey(nil)(options)
		With.EncVerifySignaturePublicKey(nil)(options)
		With.EncVerifyServerSignature(false)(options)
		With.Compression(gtp.Compression_Brotli)(options)
		With.CompressedSize(1024 * 32)(options)
		With.AutoReconnect(false)(options)
		With.AutoReconnectInterval(3 * time.Second)(options)
		With.AutoReconnectRetryTimes(100)(options)
		With.InactiveTimeout(time.Minute)(options)
		With.SendDataChanSize(0)(options)
		With.RecvDataChanSize(0, false)(options)
		With.SendEventChanSize(0)(options)
		With.RecvEventChanSize(0)(options)
		With.RecvDataHandler(nil)(options)
		With.RecvEventHandler(nil)(options)
		With.FutureTimeout(5 * time.Second)(options)
		With.AuthUserId("")(options)
		With.AuthToken("")(options)
		With.AuthExtensions(nil)(options)
		With.ZapLogger(zap.NewExample())(options)
	}
}

func (_Option) NetProtocol(p NetProtocol) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.NetProtocol = p
	}
}

func (_Option) TCPNoDelay(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPNoDelay = b
	}
}

func (_Option) TCPQuickAck(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPQuickAck = b
	}
}

func (_Option) TCPRecvBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPRecvBuf = size
	}
}

func (_Option) TCPSendBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPSendBuf = size
	}
}

func (_Option) TCPLinger(sec *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPLinger = sec
	}
}

func (_Option) WebSocketOrigin(origin string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.WebSocketOrigin = origin
	}
}

func (_Option) TLSConfig(tlsConfig *tls.Config) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TLSConfig = tlsConfig
	}
}

func (_Option) IOTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.IOTimeout = d
	}
}

func (_Option) IORetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.IORetryTimes = times
	}
}

func (_Option) IOBufferCap(cap int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.IOBufferCap = cap
	}
}

func (_Option) DecoderMsgCreator(mc gtp.IMsgCreator) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if mc == nil {
			panic(fmt.Errorf("%w: option DecoderMsgCreator can't be assigned to nil", core.ErrArgs))
		}
		options.DecoderMsgCreator = mc
	}
}

func (_Option) EncCipherSuite(cs gtp.CipherSuite) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncCipherSuite = cs
	}
}

func (_Option) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

func (_Option) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

func (_Option) EncVerifyServerSignature(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifyServerSignature = b
	}
}

func (_Option) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

func (_Option) Compression(c gtp.Compression) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.Compression = c
	}
}

func (_Option) CompressedSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.CompressedSize = size
	}
}

func (_Option) AutoReconnect(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnect = b
	}
}

func (_Option) AutoReconnectInterval(dur time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnectInterval = dur
	}
}

func (_Option) AutoReconnectRetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnectRetryTimes = times
	}
}

func (_Option) InactiveTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.InactiveTimeout = d
	}
}

func (_Option) SendDataChanSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size > 0 {
			options.SendDataChan = make(chan binaryutil.RecycleBytes, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (_Option) RecvDataChanSize(size int, recyclable bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan binaryutil.RecycleBytes, size)
		} else {
			options.RecvDataChan = nil
		}
		options.RecvDataChanRecyclable = recyclable
	}
}

func (_Option) SendEventChanSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.Event[gtp.MsgReader], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (_Option) RecvEventChanSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}

func (_Option) RecvDataHandler(handler RecvDataHandler) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.RecvDataHandler = handler
	}
}

func (_Option) RecvEventHandler(handler RecvEventHandler) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.RecvEventHandler = handler
	}
}

func (_Option) FutureTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.FutureTimeout = d
	}
}

func (_Option) AuthUserId(userId string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthUserId = userId
	}
}

func (_Option) AuthToken(token string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthToken = token
	}
}

func (_Option) AuthExtensions(extensions []byte) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthExtensions = extensions
	}
}

func (_Option) ZapLogger(logger *zap.Logger) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if logger == nil {
			panic(fmt.Errorf("%w: option ZapLogger can't be assigned to nil", core.ErrArgs))
		}
		options.ZapLogger = logger
	}
}

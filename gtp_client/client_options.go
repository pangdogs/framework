package gtp_client

import (
	"crypto"
	"crypto/tls"
	"go.uber.org/zap"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/transport"
	"time"
)

type Option struct{}

type (
	RecvDataHandler  = func(client *Client, data []byte) error                    // 客户端接收的数据的处理器
	RecvEventHandler = func(client *Client, event transport.Event[gtp.Msg]) error // 客户端接收的自定义事件的处理器
)

type ClientOptions struct {
	TLSConfig                   *tls.Config                   // TLS配置，nil表示不使用TLS加密链路
	TCPNoDelay                  *bool                         // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                 *bool                         // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                  *int                          // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                  *int                          // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                   *int                          // TCP的PLinger选项，nil表示使用系统默认值
	IOTimeout                   time.Duration                 // 网络io超时时间
	IORetryTimes                int                           // 网络io超时后的重试次数
	IOBufferCap                 int                           // 网络io缓存容量（字节）
	DecoderMsgCreator           codec.IMsgCreator             // 消息包解码器的消息构建器
	EncCipherSuite              gtp.CipherSuite               // 加密通信中的密码学套件
	EncSignatureAlgorithm       gtp.SignatureAlgorithm        // 加密通信中的签名算法
	EncSignaturePrivateKey      crypto.PrivateKey             // 加密通信中，签名用的私钥
	EncVerifyServerSignature    bool                          // 加密通信中，是否验证服务端签名
	EncVerifySignaturePublicKey crypto.PublicKey              // 加密通信中，验证服务端签名用的公钥
	Compression                 gtp.Compression               // 通信中的压缩函数
	CompressedSize              int                           // 通信中启用压缩阀值（字节），<=0表示不开启
	AutoReconnect               bool                          // 开启自动重连
	AutoReconnectInterval       time.Duration                 // 自动重连的时间间隔
	AutoReconnectRetryTimes     int                           // 自动重连的重试次数，<=0表示无限重试
	InactiveTimeout             time.Duration                 // 连接不活跃后的超时时间，开启自动重连后无效
	SendDataChan                chan []byte                   // 发送数据的channel
	RecvDataChan                chan []byte                   // 接收数据的channel
	SendEventChan               chan transport.Event[gtp.Msg] // 发送自定义事件的channel
	RecvEventChan               chan transport.Event[gtp.Msg] // 接收自定义事件的channel
	RecvDataHandlers            []RecvDataHandler             // 接收的数据的处理器列表
	RecvEventHandlers           []RecvEventHandler            // 接收的自定义事件的处理器列表
	PromiseTimeout              time.Duration                 // 异步编程模型承诺（Promise）请求超时时间
	AuthToken                   string                        // 鉴权token
	AuthExtensions              []byte                        // 鉴权extensions
	ZapLogger                   *zap.Logger                   // zap日志
}

type ClientOption func(options *ClientOptions)

func (Option) Default() ClientOption {
	return func(options *ClientOptions) {
		Option{}.TLSConfig(nil)(options)
		Option{}.TCPNoDelay(nil)(options)
		Option{}.TCPQuickAck(nil)(options)
		Option{}.TCPRecvBuf(nil)(options)
		Option{}.TCPSendBuf(nil)(options)
		Option{}.TCPLinger(nil)(options)
		Option{}.IOTimeout(3 * time.Second)(options)
		Option{}.IORetryTimes(3)(options)
		Option{}.IOBufferCap(1024 * 128)(options)
		Option{}.DecoderMsgCreator(codec.DefaultMsgCreator())(options)
		Option{}.EncCipherSuite(gtp.CipherSuite{
			SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
			SymmetricEncryption: gtp.SymmetricEncryption_AES,
			BlockCipherMode:     gtp.BlockCipherMode_CTR,
			PaddingMode:         gtp.PaddingMode_None,
			MACHash:             gtp.Hash_Fnv1a32,
		})(options)
		Option{}.EncSignatureAlgorithm(gtp.SignatureAlgorithm{
			AsymmetricEncryption: gtp.AsymmetricEncryption_None,
			PaddingMode:          gtp.PaddingMode_None,
			Hash:                 gtp.Hash_None,
		})(options)
		Option{}.EncSignaturePrivateKey(nil)(options)
		Option{}.EncVerifySignaturePublicKey(nil)(options)
		Option{}.EncVerifyServerSignature(false)(options)
		Option{}.Compression(gtp.Compression_Brotli)(options)
		Option{}.CompressedSize(1024 * 32)(options)
		Option{}.AutoReconnect(false)(options)
		Option{}.AutoReconnectInterval(3 * time.Second)(options)
		Option{}.AutoReconnectRetryTimes(100)(options)
		Option{}.InactiveTimeout(60 * time.Second)(options)
		Option{}.SendDataChanSize(0)(options)
		Option{}.RecvDataChanSize(0)(options)
		Option{}.SendEventSize(0)(options)
		Option{}.RecvEventSize(0)(options)
		Option{}.RecvDataHandlers(nil)(options)
		Option{}.RecvEventHandlers(nil)(options)
		Option{}.PromiseTimeout(10 * time.Second)(options)
		Option{}.AuthToken("")(options)
		Option{}.AuthExtensions(nil)(options)
		Option{}.ZapLogger(zap.NewExample())(options)
	}
}

func (Option) TLSConfig(tlsConfig *tls.Config) ClientOption {
	return func(options *ClientOptions) {
		options.TLSConfig = tlsConfig
	}
}

func (Option) TCPNoDelay(b *bool) ClientOption {
	return func(options *ClientOptions) {
		options.TCPNoDelay = b
	}
}

func (Option) TCPQuickAck(b *bool) ClientOption {
	return func(options *ClientOptions) {
		options.TCPQuickAck = b
	}
}

func (Option) TCPRecvBuf(size *int) ClientOption {
	return func(options *ClientOptions) {
		options.TCPRecvBuf = size
	}
}

func (Option) TCPSendBuf(size *int) ClientOption {
	return func(options *ClientOptions) {
		options.TCPSendBuf = size
	}
}

func (Option) TCPLinger(sec *int) ClientOption {
	return func(options *ClientOptions) {
		options.TCPLinger = sec
	}
}

func (Option) IOTimeout(d time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.IOTimeout = d
	}
}

func (Option) IORetryTimes(times int) ClientOption {
	return func(options *ClientOptions) {
		options.IORetryTimes = times
	}
}

func (Option) IOBufferCap(cap int) ClientOption {
	return func(options *ClientOptions) {
		options.IOBufferCap = cap
	}
}

func (Option) DecoderMsgCreator(mc codec.IMsgCreator) ClientOption {
	return func(options *ClientOptions) {
		if mc == nil {
			panic("option DecoderMsgCreator can't be assigned to nil")
		}
		options.DecoderMsgCreator = mc
	}
}

func (Option) EncCipherSuite(cs gtp.CipherSuite) ClientOption {
	return func(options *ClientOptions) {
		options.EncCipherSuite = cs
	}
}

func (Option) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) ClientOption {
	return func(options *ClientOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

func (Option) EncSignaturePrivateKey(priv crypto.PrivateKey) ClientOption {
	return func(options *ClientOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

func (Option) EncVerifyServerSignature(b bool) ClientOption {
	return func(options *ClientOptions) {
		options.EncVerifyServerSignature = b
	}
}

func (Option) EncVerifySignaturePublicKey(pub crypto.PublicKey) ClientOption {
	return func(options *ClientOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

func (Option) Compression(c gtp.Compression) ClientOption {
	return func(options *ClientOptions) {
		options.Compression = c
	}
}

func (Option) CompressedSize(size int) ClientOption {
	return func(options *ClientOptions) {
		options.CompressedSize = size
	}
}

func (Option) AutoReconnect(b bool) ClientOption {
	return func(options *ClientOptions) {
		options.AutoReconnect = b
	}
}

func (Option) AutoReconnectInterval(dur time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.AutoReconnectInterval = dur
	}
}

func (Option) AutoReconnectRetryTimes(times int) ClientOption {
	return func(options *ClientOptions) {
		options.AutoReconnectRetryTimes = times
	}
}

func (Option) InactiveTimeout(d time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.InactiveTimeout = d
	}
}

func (Option) SendDataChanSize(size int) ClientOption {
	return func(options *ClientOptions) {
		if size > 0 {
			options.SendDataChan = make(chan []byte, size)
		} else {
			options.SendDataChan = nil
		}
	}
}

func (Option) RecvDataChanSize(size int) ClientOption {
	return func(options *ClientOptions) {
		if size > 0 {
			options.RecvDataChan = make(chan []byte, size)
		} else {
			options.RecvDataChan = nil
		}
	}
}

func (Option) SendEventSize(size int) ClientOption {
	return func(options *ClientOptions) {
		if size > 0 {
			options.SendEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.SendEventChan = nil
		}
	}
}

func (Option) RecvEventSize(size int) ClientOption {
	return func(options *ClientOptions) {
		if size > 0 {
			options.RecvEventChan = make(chan transport.Event[gtp.Msg], size)
		} else {
			options.RecvEventChan = nil
		}
	}
}

func (Option) RecvDataHandlers(handlers ...RecvDataHandler) ClientOption {
	return func(options *ClientOptions) {
		options.RecvDataHandlers = handlers
	}
}

func (Option) RecvEventHandlers(handlers ...RecvEventHandler) ClientOption {
	return func(options *ClientOptions) {
		options.RecvEventHandlers = handlers
	}
}

func (Option) PromiseTimeout(d time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.PromiseTimeout = d
	}
}

func (Option) AuthToken(token string) ClientOption {
	return func(options *ClientOptions) {
		options.AuthToken = token
	}
}

func (Option) AuthExtensions(extensions []byte) ClientOption {
	return func(options *ClientOptions) {
		options.AuthExtensions = extensions
	}
}

func (Option) ZapLogger(logger *zap.Logger) ClientOption {
	return func(options *ClientOptions) {
		if logger == nil {
			panic("option ZapLogger can't be assigned to nil")
		}
		options.ZapLogger = logger
	}
}

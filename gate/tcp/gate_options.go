package tcp

import (
	"crypto/tls"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"net"
	"time"
)

type WithOption struct{}

type GateOptions struct {
	Endpoints               []string                     // 所有监听地址
	TLSConfig               *tls.Config                  // TLS配置，nil表示不使用TLS加密链路
	TCPNoDelay              *bool                        // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck             *bool                        // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf              *int                         // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf              *int                         // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger               *int                         // TCP的PLinger选项，nil表示使用系统默认值
	Timeout                 time.Duration                // 消息收发超时时间
	MsgCreator              codec.IMsgCreator            // 消息构建器
	AgreeCliProposal        bool                         // 是否同意使用客户端建议的加密与压缩方案
	CipherSuite             transport.CipherSuite        // 密码学套件
	CompressionMethod       transport.CompressionMethod  // 压缩函数
	ECDHENamedCurve         transport.NamedCurve         // ECDHE交换秘钥时使用的曲线类型
	ECDHESignatureAlgorithm transport.SignatureAlgorithm // ECDHE交换秘钥时使用的签名算法
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
		WithOption{}.MsgCreator(codec.DefaultMsgCreator())(options)
		WithOption{}.AgreeCliProposal(false)(options)
		WithOption{}.CipherSuite(transport.CipherSuite{})(options)
		WithOption{}.CompressionMethod(transport.CompressionMethod_None)(options)
		WithOption{}.ECDHENamedCurve(transport.NamedCurve_None)(options)
		WithOption{}.ECDHESignatureAlgorithm(transport.SignatureAlgorithm{})(options)
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

func (WithOption) MsgCreator(mc codec.IMsgCreator) GateOption {
	return func(options *GateOptions) {
		if mc == nil {
			panic("option MsgCreator can't be assigned to nil")
		}
		options.MsgCreator = mc
	}
}

func (WithOption) AgreeCliProposal(b bool) GateOption {
	return func(options *GateOptions) {
		options.AgreeCliProposal = b
	}
}

func (WithOption) CipherSuite(cs transport.CipherSuite) GateOption {
	return func(options *GateOptions) {
		options.CipherSuite = cs
	}
}

func (WithOption) CompressionMethod(cm transport.CompressionMethod) GateOption {
	return func(options *GateOptions) {
		options.CompressionMethod = cm
	}
}

func (WithOption) ECDHENamedCurve(nc transport.NamedCurve) GateOption {
	return func(options *GateOptions) {
		options.ECDHENamedCurve = nc
	}
}

func (WithOption) ECDHESignatureAlgorithm(sa transport.SignatureAlgorithm) GateOption {
	return func(options *GateOptions) {
		options.ECDHESignatureAlgorithm = sa
	}
}

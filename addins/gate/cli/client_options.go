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

package cli

import (
	"crypto"
	"crypto/tls"
	"net/url"
	"strings"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gtp"
	"go.uber.org/zap"
)

// NetProtocol 网络协议
type NetProtocol int32

const (
	TCP NetProtocol = iota
	WebSocket
)

// ClientOptions 客户端所有选项
type ClientOptions struct {
	NetProtocol                 NetProtocol            // 使用的网络协议（TCP/WebSocket）
	TCPNoDelay                  *bool                  // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                 *bool                  // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                  *int                   // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                  *int                   // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                   *int                   // TCP的PLinger选项，nil表示使用系统默认值
	WebSocketOrigin             string                 // WebSocket的Origin地址，不填将会自动生成
	TLSConfig                   *tls.Config            // TLS配置，nil表示不使用TLS加密链路
	IOTimeout                   time.Duration          // 网络io超时时间
	IORetryTimes                int                    // 网络io超时后的重试次数
	IOBufferCap                 int                    // 网络io缓存容量（字节）
	MsgCreator                  gtp.IMsgCreator        // 消息包解码器的消息构建器
	EncCipherSuite              gtp.CipherSuite        // 加密通信中的密码学套件
	EncSignatureAlgorithm       gtp.SignatureAlgorithm // 加密通信中的签名算法
	EncSignaturePrivateKey      crypto.PrivateKey      // 加密通信中，签名用的私钥
	EncVerifyServerSignature    bool                   // 加密通信中，是否验证服务端签名
	EncVerifySignaturePublicKey crypto.PublicKey       // 加密通信中，验证服务端签名用的公钥
	Compression                 gtp.Compression        // 通信中的压缩函数
	CompressionThreshold        int                    // 通信中启用压缩阀值（字节），<=0表示不开启
	MaxUncompressedSize         int                    // 通信中最大解压缩大小，用于防御压缩包炸弹
	AutoReconnect               bool                   // 开启自动重连
	AutoReconnectInterval       time.Duration          // 自动重连的时间间隔
	AutoReconnectRetryTimes     int                    // 自动重连的重试次数，<=0表示无限重试
	InactiveTimeout             time.Duration          // 连接不活跃后的超时时间，开启自动重连后无效
	FutureTimeout               time.Duration          // 异步模型Future超时时间
	AuthUserId                  string                 // 鉴权userid
	AuthToken                   string                 // 鉴权token
	AuthExtensions              []byte                 // 鉴权extensions
	AutoRecover                 bool                   // panic时是否自动恢复
	ReportError                 chan error             // 在开启panic时自动恢复时，将会恢复并将错误写入此error channel
	DataListenerInboxSize       int                    // 数据监听器inbox缓存大小
	EventListenerInboxSize      int                    // 事件监听器inbox缓存大小
	Logger                      *zap.Logger            // 日志
}

var With _ClientOption

type _ClientOption struct{}

// Default 默认选项
func (_ClientOption) Default() option.Setting[ClientOptions] {
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
		With.IOBufferCap(128 * 1024)(options)
		With.MsgCreator(gtp.DefaultMsgCreator())(options)
		With.EncCipherSuite(gtp.CipherSuite{
			SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
			SymmetricEncryption: gtp.SymmetricEncryption_XChaCha20_Poly1305,
			BlockCipherMode:     gtp.BlockCipherMode_None,
			PaddingMode:         gtp.PaddingMode_None,
			HMAC:                gtp.Hash_None,
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
		With.CompressionThreshold(64 * 1024)(options)
		With.MaxUncompressedSize(64 * 1024 * 1024)(options)
		With.AutoReconnect(false)(options)
		With.AutoReconnectInterval(3 * time.Second)(options)
		With.AutoReconnectRetryTimes(100)(options)
		With.InactiveTimeout(time.Minute)(options)
		With.FutureTimeout(5 * time.Second)(options)
		With.AuthUserId("")(options)
		With.AuthToken("")(options)
		With.AuthExtensions(nil)(options)
		With.PanicHandling(false, nil)(options)
		With.DataListenerInboxSize(128)(options)
		With.EventListenerInboxSize(128)(options)
		With.Logger(nil)(options)
	}
}

// NetProtocol 设置使用的网络协议（TCP/WebSocket）
func (_ClientOption) NetProtocol(p NetProtocol) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.NetProtocol = p
	}
}

// TCPNoDelay 设置TCP的NoDelay选项，nil表示使用系统默认值
func (_ClientOption) TCPNoDelay(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPNoDelay = b
	}
}

// TCPQuickAck 设置TCP的QuickAck选项，nil表示使用系统默认值
func (_ClientOption) TCPQuickAck(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPQuickAck = b
	}
}

// TCPRecvBuf 设置TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
func (_ClientOption) TCPRecvBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPRecvBuf = size
	}
}

// TCPSendBuf 设置TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
func (_ClientOption) TCPSendBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPSendBuf = size
	}
}

// TCPLinger 设置TCP的Linger选项，nil表示使用系统默认值
func (_ClientOption) TCPLinger(sec *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPLinger = sec
	}
}

// WebSocketOrigin 设置WebSocket的Origin地址，不填将会自动生成
func (_ClientOption) WebSocketOrigin(origin string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if origin != "" {
			url, err := url.Parse(origin)
			if err != nil {
				exception.Panicf("cli: %w: %w", core.ErrArgs, err)
			}
			switch strings.ToLower(url.Scheme) {
			case "http", "https", "ws", "wss":
			default:
				exception.Panicf("cli: %w: option WebSocketOrigin has unsupported scheme %q", core.ErrArgs, url.Scheme)
			}
			if url.Host == "" {
				exception.Panicf("cli: %w: option WebSocketOrigin host can't be empty", core.ErrArgs)
			}
		}
		options.WebSocketOrigin = origin
	}
}

// TLSConfig 设置TLS配置，nil表示不使用TLS加密链路
func (_ClientOption) TLSConfig(tlsConfig *tls.Config) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TLSConfig = tlsConfig
	}
}

// IOTimeout 设置网络io超时时间
func (_ClientOption) IOTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 100*time.Millisecond {
			exception.Panicf("cli: %w: option IOTimeout must be >= 0.1 seconds", core.ErrArgs)
		}
		options.IOTimeout = d
	}
}

// IORetryTimes 设置网络io超时后的重试次数
func (_ClientOption) IORetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if times < 0 {
			exception.Panicf("cli: %w: option IORetryTimes must be >= 0", core.ErrArgs)
		}
		options.IORetryTimes = times
	}
}

// IOBufferCap 设置网络io缓存容量（字节）
func (_ClientOption) IOBufferCap(cap int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if cap < 1024 {
			exception.Panicf("cli: %w: option IOBufferCap must be >= 1024 bytes", core.ErrArgs)
		}
		options.IOBufferCap = cap
	}
}

// MsgCreator 设置消息包解码器的消息构建器
func (_ClientOption) MsgCreator(mc gtp.IMsgCreator) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if mc == nil {
			exception.Panicf("cli: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}

// EncCipherSuite 设置加密通信中的密码学套件
func (_ClientOption) EncCipherSuite(cs gtp.CipherSuite) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncCipherSuite = cs
	}
}

// EncSignatureAlgorithm 设置加密通信中的签名算法
func (_ClientOption) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

// EncSignaturePrivateKey 设置加密通信中，签名用的私钥
func (_ClientOption) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

// EncVerifyServerSignature 设置加密通信中，是否验证服务端签名
func (_ClientOption) EncVerifyServerSignature(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifyServerSignature = b
	}
}

// EncVerifySignaturePublicKey 设置加密通信中，验证服务端签名用的公钥
func (_ClientOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

// Compression 设置通信中的压缩函数
func (_ClientOption) Compression(c gtp.Compression) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.Compression = c
	}
}

// CompressionThreshold 设置通信中启用压缩阀值（字节），<=0表示不开启
func (_ClientOption) CompressionThreshold(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.CompressionThreshold = size
	}
}

// MaxUncompressedSize 通信中最大解压缩大小，用于防御压缩包炸弹
func (_ClientOption) MaxUncompressedSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.MaxUncompressedSize = size
	}
}

// AutoReconnect 设置开启自动重连
func (_ClientOption) AutoReconnect(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnect = b
	}
}

// AutoReconnectInterval 设置自动重连的时间间隔
func (_ClientOption) AutoReconnectInterval(dur time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if dur < 0 {
			exception.Panicf("cli: %w: option AutoReconnectInterval must be >= 0 seconds", core.ErrArgs)
		}
		options.AutoReconnectInterval = dur
	}
}

// AutoReconnectRetryTimes 设置自动重连的重试次数，<=0表示无限重试
func (_ClientOption) AutoReconnectRetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnectRetryTimes = times
	}
}

// InactiveTimeout 设置连接不活跃后的超时时间，开启自动重连后无效
func (_ClientOption) InactiveTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 0 {
			exception.Panicf("cli: %w: option InactiveTimeout must be >= 0 seconds", core.ErrArgs)
		}
		options.InactiveTimeout = d
	}
}

// FutureTimeout 设置异步模型Future超时时间
func (_ClientOption) FutureTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("cli: %w: option FutureTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.FutureTimeout = d
	}
}

// AuthUserId 设置鉴权userid
func (_ClientOption) AuthUserId(userId string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthUserId = userId
	}
}

// AuthToken 设置鉴权token
func (_ClientOption) AuthToken(token string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthToken = token
	}
}

// AuthExtensions 设置鉴权extensions
func (_ClientOption) AuthExtensions(extensions []byte) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthExtensions = extensions
	}
}

// PanicHandling 设置panic时的处理方式
func (_ClientOption) PanicHandling(autoRecover bool, reportError chan error) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoRecover = autoRecover
		options.ReportError = reportError
	}
}

// DataListenerInboxSize 设置数据监听器inbox缓存大小
func (_ClientOption) DataListenerInboxSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size <= 0 {
			exception.Panicf("cli: %w: option DataListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.DataListenerInboxSize = size
	}
}

// EventListenerInboxSize 设置事件监听器inbox缓存大小
func (_ClientOption) EventListenerInboxSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size <= 0 {
			exception.Panicf("cli: %w: option EventListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.EventListenerInboxSize = size
	}
}

// Logger 设置日志
func (_ClientOption) Logger(logger *zap.Logger) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.Logger = logger
	}
}

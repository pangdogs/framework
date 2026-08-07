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

// NetProtocol 选择客户端建立底层连接的协议。
type NetProtocol int32

const (
	// TCP 使用原生 TCP 连接。
	TCP NetProtocol = iota
	// WebSocket 使用二进制 WebSocket 连接。
	WebSocket
)

// ClientOptions 配置连接协议、GTP 协商、重连、监听器和日志行为。
type ClientOptions struct {
	NetProtocol                 NetProtocol            // NetProtocol 选择 TCP 或 WebSocket。
	TCPNoDelay                  *bool                  // TCPNoDelay 为 nil 时沿用系统默认值。
	TCPQuickAck                 *bool                  // TCPQuickAck 为 nil 时沿用系统默认值。
	TCPRecvBuf                  *int                   // TCPRecvBuf 是接收缓冲区字节数；nil 沿用系统默认值。
	TCPSendBuf                  *int                   // TCPSendBuf 是发送缓冲区字节数；nil 沿用系统默认值。
	TCPLinger                   *int                   // TCPLinger 是关闭等待秒数；nil 沿用系统默认值。
	WebSocketOrigin             string                 // WebSocketOrigin 为空时根据 endpoint 和用户 ID 生成。
	TLSConfig                   *tls.Config            // TLSConfig 非 nil 时为 TCP 或安全 WebSocket 启用 TLS。
	IOTimeout                   time.Duration          // IOTimeout 是单次网络 I/O 的超时。
	IORetryTimes                int                    // IORetryTimes 是 I/O 超时后的重试次数。
	IOBufferCap                 int                    // IOBufferCap 是断线重连时保留的发送数据字节上限。
	MsgCreator                  gtp.IMsgCreator        // MsgCreator 用于按消息 ID 创建 GTP 解码目标。
	EncCipherSuite              gtp.CipherSuite        // EncCipherSuite 是客户端提出的密码套件。
	EncSignatureAlgorithm       gtp.SignatureAlgorithm // EncSignatureAlgorithm 是客户端握手签名算法。
	EncSignaturePrivateKey      crypto.PrivateKey      // EncSignaturePrivateKey 是客户端握手签名私钥。
	EncVerifyServerSignature    bool                   // EncVerifyServerSignature 要求验证服务端握手签名。
	EncVerifySignaturePublicKey crypto.PublicKey       // EncVerifySignaturePublicKey 是服务端签名验证公钥。
	Compression                 gtp.Compression        // Compression 是客户端提出的压缩算法。
	CompressionThreshold        int                    // CompressionThreshold 是启用压缩的字节阈值；小于等于 0 时禁用。
	MaxUncompressedSize         int                    // MaxUncompressedSize 限制解压后负载，防御压缩炸弹。
	MaxPacketSize               int                    // MaxPacketSize 限制单个 GTP 包大小。
	AutoReconnect               bool                   // AutoReconnect 在连接失活后自动迁移会话连接。
	AutoReconnectInterval       time.Duration          // AutoReconnectInterval 是相邻重连尝试的间隔。
	AutoReconnectRetryTimes     int                    // AutoReconnectRetryTimes 小于等于 0 时无限重试。
	InactiveTimeout             time.Duration          // InactiveTimeout 是未启用自动重连时的失活等待时间。
	FutureTimeout               time.Duration          // FutureTimeout 是时间探测等关联请求的默认超时。
	AuthUserId                  string                 // AuthUserId 是握手提交的用户 ID。
	AuthToken                   string                 // AuthToken 是握手提交的鉴权令牌。
	AuthExtensions              []byte                 // AuthExtensions 是握手提交的扩展数据。
	AutoRecover                 bool                   // AutoRecover 控制监听器 panic 是否自动恢复。
	ReportError                 chan error             // ReportError 接收自动恢复的 panic 错误。
	DataListenerInboxSize       int                    // DataListenerInboxSize 是每个数据监听器的收件箱容量。
	EventListenerInboxSize      int                    // EventListenerInboxSize 是每个事件监听器的收件箱容量。
	Logger                      *zap.Logger            // Logger 是客户端日志器；nil 时不输出日志。
}

// With 提供 gate 客户端的 Option 构造方法。
var With _ClientOption

type _ClientOption struct{}

// Default 返回 TCP、三秒 I/O 超时、关闭自动重连及默认 GTP 安全参数。
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
		With.MaxUncompressedSize(128 * 1024 * 1024)(options)
		With.MaxPacketSize(64 * 1024 * 1024)(options)
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

// NetProtocol 设置底层连接协议。
func (_ClientOption) NetProtocol(p NetProtocol) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.NetProtocol = p
	}
}

// TCPNoDelay 设置 TCP_NODELAY；nil 表示使用系统默认值。
func (_ClientOption) TCPNoDelay(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPNoDelay = b
	}
}

// TCPQuickAck 设置 TCP_QUICKACK；nil 表示使用系统默认值。
func (_ClientOption) TCPQuickAck(b *bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPQuickAck = b
	}
}

// TCPRecvBuf 设置 TCP 接收缓冲区字节数；nil 表示使用系统默认值。
func (_ClientOption) TCPRecvBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPRecvBuf = size
	}
}

// TCPSendBuf 设置 TCP 发送缓冲区字节数；nil 表示使用系统默认值。
func (_ClientOption) TCPSendBuf(size *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPSendBuf = size
	}
}

// TCPLinger 设置 TCP 关闭等待秒数；nil 表示使用系统默认值。
func (_ClientOption) TCPLinger(sec *int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TCPLinger = sec
	}
}

// WebSocketOrigin 设置 WebSocket Origin；空字符串表示根据 endpoint 和用户 ID 生成。
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

// TLSConfig 设置底层 TLS 配置；nil 表示不额外启用 TLS。
func (_ClientOption) TLSConfig(tlsConfig *tls.Config) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.TLSConfig = tlsConfig
	}
}

// IOTimeout 设置单次网络 I/O 超时，必须不少于 100 毫秒。
func (_ClientOption) IOTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 100*time.Millisecond {
			exception.Panicf("cli: %w: option IOTimeout must be >= 0.1 seconds", core.ErrArgs)
		}
		options.IOTimeout = d
	}
}

// IORetryTimes 设置网络 I/O 超时后的重试次数，必须大于等于 0。
func (_ClientOption) IORetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if times < 0 {
			exception.Panicf("cli: %w: option IORetryTimes must be >= 0", core.ErrArgs)
		}
		options.IORetryTimes = times
	}
}

// IOBufferCap 设置断线重连时保留的发送数据字节上限，必须不少于 1024。
func (_ClientOption) IOBufferCap(cap int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if cap < 1024 {
			exception.Panicf("cli: %w: option IOBufferCap must be >= 1024 bytes", core.ErrArgs)
		}
		options.IOBufferCap = cap
	}
}

// MsgCreator 设置 GTP 消息构建器，不得为 nil。
func (_ClientOption) MsgCreator(mc gtp.IMsgCreator) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if mc == nil {
			exception.Panicf("cli: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}

// EncCipherSuite 设置客户端向服务端提出的密码套件。
func (_ClientOption) EncCipherSuite(cs gtp.CipherSuite) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncCipherSuite = cs
	}
}

// EncSignatureAlgorithm 设置客户端握手签名算法。
func (_ClientOption) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

// EncSignaturePrivateKey 设置客户端握手签名私钥。
func (_ClientOption) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

// EncVerifyServerSignature 设置是否验证服务端握手签名。
func (_ClientOption) EncVerifyServerSignature(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifyServerSignature = b
	}
}

// EncVerifySignaturePublicKey 设置验证服务端握手签名使用的公钥。
func (_ClientOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

// Compression 设置客户端向服务端提出的压缩算法。
func (_ClientOption) Compression(c gtp.Compression) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.Compression = c
	}
}

// CompressionThreshold 设置启用压缩的字节阈值；小于等于 0 时禁用压缩。
func (_ClientOption) CompressionThreshold(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.CompressionThreshold = size
	}
}

// MaxUncompressedSize 设置解压后负载字节上限，用于防御压缩炸弹。
func (_ClientOption) MaxUncompressedSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.MaxUncompressedSize = size
	}
}

// MaxPacketSize 设置单个 GTP 包的字节上限。
func (_ClientOption) MaxPacketSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.MaxPacketSize = size
	}
}

// AutoReconnect 设置连接失活后是否自动迁移会话连接。
func (_ClientOption) AutoReconnect(b bool) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnect = b
	}
}

// AutoReconnectInterval 设置相邻重连尝试的间隔，必须大于等于 0。
func (_ClientOption) AutoReconnectInterval(dur time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if dur < 0 {
			exception.Panicf("cli: %w: option AutoReconnectInterval must be >= 0 seconds", core.ErrArgs)
		}
		options.AutoReconnectInterval = dur
	}
}

// AutoReconnectRetryTimes 设置自动重连尝试次数；小于等于 0 时无限重试。
func (_ClientOption) AutoReconnectRetryTimes(times int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoReconnectRetryTimes = times
	}
}

// InactiveTimeout 设置未启用自动重连时的失活等待时间，必须大于等于 0。
func (_ClientOption) InactiveTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 0 {
			exception.Panicf("cli: %w: option InactiveTimeout must be >= 0 seconds", core.ErrArgs)
		}
		options.InactiveTimeout = d
	}
}

// FutureTimeout 设置关联请求 Future 的默认超时，必须不少于 300 毫秒。
func (_ClientOption) FutureTimeout(d time.Duration) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("cli: %w: option FutureTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.FutureTimeout = d
	}
}

// AuthUserId 设置握手提交的用户 ID。
func (_ClientOption) AuthUserId(userId string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthUserId = userId
	}
}

// AuthToken 设置握手提交的鉴权令牌。
func (_ClientOption) AuthToken(token string) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthToken = token
	}
}

// AuthExtensions 设置握手提交的扩展数据；切片不会复制。
func (_ClientOption) AuthExtensions(extensions []byte) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AuthExtensions = extensions
	}
}

// PanicHandling 设置监听器是否自动恢复 panic 以及错误报告通道。
func (_ClientOption) PanicHandling(autoRecover bool, reportError chan error) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.AutoRecover = autoRecover
		options.ReportError = reportError
	}
}

// DataListenerInboxSize 设置每个数据监听器的收件箱容量，必须大于 0。
func (_ClientOption) DataListenerInboxSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size <= 0 {
			exception.Panicf("cli: %w: option DataListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.DataListenerInboxSize = size
	}
}

// EventListenerInboxSize 设置每个事件监听器的收件箱容量，必须大于 0。
func (_ClientOption) EventListenerInboxSize(size int) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		if size <= 0 {
			exception.Panicf("cli: %w: option EventListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.EventListenerInboxSize = size
	}
}

// Logger 设置客户端日志器；nil 表示使用不输出日志的实现。
func (_ClientOption) Logger(logger *zap.Logger) option.Setting[ClientOptions] {
	return func(options *ClientOptions) {
		options.Logger = logger
	}
}

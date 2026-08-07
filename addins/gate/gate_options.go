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

package gate

import (
	"crypto"
	"crypto/tls"
	"math/big"
	"net"
	"net/url"
	"strings"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/net/gtp"
	"golang.org/x/net/websocket"
)

type (
	// WebSocketAddrResolver 从 WebSocket 握手与连接信息中解析一个网络地址。
	WebSocketAddrResolver = generic.Func1[*websocket.Conn, net.Addr]
	// Authenticator 校验客户端提交的用户 ID、令牌和扩展数据；返回错误会拒绝连接。
	Authenticator = generic.Delegate5[IGate, net.Conn, string, string, []byte, error]
)

// GateOptions 配置监听端点、GTP 协商、安全限制、会话及监听器容量。
type GateOptions struct {
	TCPAddress                     string                 // TCPAddress 是 TCP 监听地址；空字符串禁用 TCP。
	TCPNoDelay                     *bool                  // TCPNoDelay 为 nil 时沿用系统默认值。
	TCPQuickAck                    *bool                  // TCPQuickAck 为 nil 时沿用系统默认值。
	TCPRecvBuf                     *int                   // TCPRecvBuf 是接收缓冲区字节数；nil 沿用系统默认值。
	TCPSendBuf                     *int                   // TCPSendBuf 是发送缓冲区字节数；nil 沿用系统默认值。
	TCPLinger                      *int                   // TCPLinger 是关闭等待秒数；nil 沿用系统默认值。
	TCPTLSConfig                   *tls.Config            // TCPTLSConfig 非 nil 时为 TCP 监听启用 TLS。
	WebSocketURL                   *url.URL               // WebSocketURL 是监听 URL；nil 禁用 WebSocket。
	WebSocketTLSConfig             *tls.Config            // WebSocketTLSConfig 用于 https/wss 监听。
	WebSocketLocalAddrResolver     WebSocketAddrResolver  // WebSocketLocalAddrResolver 解析服务端地址。
	WebSocketRemoteAddrResolver    WebSocketAddrResolver  // WebSocketRemoteAddrResolver 解析客户端地址。
	IOTimeout                      time.Duration          // IOTimeout 是单次网络 I/O 的超时。
	IORetryTimes                   int                    // IORetryTimes 是 I/O 超时后的重试次数。
	IOBufferCap                    int                    // IOBufferCap 是断线重连时保留的发送数据字节上限。
	MsgCreator                     gtp.IMsgCreator        // MsgCreator 用于按消息 ID 创建 GTP 解码目标。
	AgreeClientEncryptionProposal  bool                   // AgreeClientEncryptionProposal 允许采用客户端加密提案。
	EncCipherSuite                 gtp.CipherSuite        // EncCipherSuite 是服务端首选密码套件。
	EncNonceStep                   *big.Int               // EncNonceStep 是每次加解密后的 nonce 增量。
	EncECDHENamedCurve             gtp.NamedCurve         // EncECDHENamedCurve 是 ECDHE 密钥交换曲线。
	EncSignatureAlgorithm          gtp.SignatureAlgorithm // EncSignatureAlgorithm 是握手签名算法。
	EncSignaturePrivateKey         crypto.PrivateKey      // EncSignaturePrivateKey 是服务端握手签名私钥。
	EncVerifyClientSignature       bool                   // EncVerifyClientSignature 要求验证客户端握手签名。
	EncVerifySignaturePublicKey    crypto.PublicKey       // EncVerifySignaturePublicKey 是客户端签名验证公钥。
	AgreeClientCompressionProposal bool                   // AgreeClientCompressionProposal 允许采用客户端压缩提案。
	Compression                    gtp.Compression        // Compression 是服务端首选压缩算法。
	CompressionThreshold           int                    // CompressionThreshold 是启用压缩的字节阈值；小于等于 0 时禁用。
	MaxUncompressedSize            int                    // MaxUncompressedSize 限制解压后负载，防御压缩炸弹。
	MaxPacketSize                  int                    // MaxPacketSize 限制单个 GTP 包大小。
	AcceptTimeout                  time.Duration          // AcceptTimeout 是握手完成前的最长等待时间。
	Authenticator                  Authenticator          // Authenticator 校验客户端凭据；nil 表示不额外鉴权。
	SessionInactiveTimeout         time.Duration          // SessionInactiveTimeout 是断线后等待连接迁移的时间。
	SessionWatcherInboxSize        int                    // SessionWatcherInboxSize 是每个会话观察器的收件箱容量。
	SessionDataListenerInboxSize   int                    // SessionDataListenerInboxSize 是每个数据监听器的收件箱容量。
	SessionEventListenerInboxSize  int                    // SessionEventListenerInboxSize 是每个事件监听器的收件箱容量。
}

// With 提供 gate add-in 的 Option 构造方法。
var With _GateOption

type _GateOption struct{}

// Default 返回同时监听 TCP :9090 和 WebSocket :80 的默认设置。
func (_GateOption) Default() option.Setting[GateOptions] {
	return func(options *GateOptions) {
		With.TCPAddress("0.0.0.0:9090")(options)
		With.TCPNoDelay(nil)(options)
		With.TCPQuickAck(nil)(options)
		With.TCPRecvBuf(nil)(options)
		With.TCPSendBuf(nil)(options)
		With.TCPLinger(nil)(options)
		With.TCPTLSConfig(nil)(options)
		With.WebSocketURL("http://0.0.0.0:80")(options)
		With.WebSocketTLSConfig(nil)(options)
		With.WebSocketLocalAddrResolver(DefaultWebSocketLocalAddrResolver)(options)
		With.WebSocketRemoteAddrResolver(DefaultWebSocketRemoteAddrResolver)(options)
		With.IOTimeout(3 * time.Second)(options)
		With.IORetryTimes(3)(options)
		With.IOBufferCap(128 * 1024)(options)
		With.MsgCreator(gtp.DefaultMsgCreator())(options)
		With.AgreeClientEncryptionProposal(true)(options)
		With.EncCipherSuite(gtp.CipherSuite{
			SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
			SymmetricEncryption: gtp.SymmetricEncryption_XChaCha20_Poly1305,
			BlockCipherMode:     gtp.BlockCipherMode_None,
			PaddingMode:         gtp.PaddingMode_None,
			HMAC:                gtp.Hash_None,
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
		With.AgreeClientCompressionProposal(true)(options)
		With.Compression(gtp.Compression_Brotli)(options)
		With.CompressionThreshold(64 * 1024)(options)
		With.MaxUncompressedSize(128 * 1024 * 1024)(options)
		With.MaxPacketSize(64 * 1024 * 1024)(options)
		With.AcceptTimeout(10 * time.Second)(options)
		With.Authenticator(nil)(options)
		With.SessionInactiveTimeout(time.Minute)(options)
		With.SessionWatcherInboxSize(256 * 1024)(options)
		With.SessionDataListenerInboxSize(128)(options)
		With.SessionEventListenerInboxSize(128)(options)
	}
}

// TCPAddress 设置 TCP 监听地址并校验 host:port 格式；空字符串禁用 TCP。
func (_GateOption) TCPAddress(addr string) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if addr != "" {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				exception.Panicf("gate: %w: %w", core.ErrArgs, err)
			}
		}
		options.TCPAddress = addr
	}
}

// TCPNoDelay 设置 TCP_NODELAY；nil 表示使用系统默认值。
func (_GateOption) TCPNoDelay(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPNoDelay = b
	}
}

// TCPQuickAck 设置 TCP_QUICKACK；nil 表示使用系统默认值。
func (_GateOption) TCPQuickAck(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPQuickAck = b
	}
}

// TCPRecvBuf 设置 TCP 接收缓冲区字节数；nil 表示使用系统默认值。
func (_GateOption) TCPRecvBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPRecvBuf = size
	}
}

// TCPSendBuf 设置 TCP 发送缓冲区字节数；nil 表示使用系统默认值。
func (_GateOption) TCPSendBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPSendBuf = size
	}
}

// TCPLinger 设置 TCP 关闭等待秒数；nil 表示使用系统默认值。
func (_GateOption) TCPLinger(sec *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPLinger = sec
	}
}

// TCPTLSConfig 设置 TCP TLS 配置；nil 表示不启用 TLS。
func (_GateOption) TCPTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPTLSConfig = tlsConfig
	}
}

// WebSocketURL 设置 WebSocket 监听 URL；空字符串禁用 WebSocket。
// 支持 http、https、ws 和 wss，未提供路径时使用根路径。
func (_GateOption) WebSocketURL(raw string) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if raw == "" {
			options.WebSocketURL = nil
			return
		}
		url, err := url.Parse(raw)
		if err != nil {
			exception.Panicf("gate: %w: %w", core.ErrArgs, err)
		}
		switch strings.ToLower(url.Scheme) {
		case "http", "https", "ws", "wss":
		default:
			exception.Panicf("gate: %w: option WebSocketURL has unsupported scheme %q", core.ErrArgs, url.Scheme)
		}
		if url.Host == "" {
			exception.Panicf("gate: %w: option WebSocketURL host can't be empty", core.ErrArgs)
		}
		if url.Path == "" {
			url.Path = "/"
		}
		options.WebSocketURL = url
	}
}

// WebSocketTLSConfig 设置 https/wss 监听使用的 TLS 配置。
func (_GateOption) WebSocketTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.WebSocketTLSConfig = tlsConfig
	}
}

// WebSocketLocalAddrResolver 设置 WebSocket 服务端地址解析器，不得为 nil。
func (_GateOption) WebSocketLocalAddrResolver(resolver WebSocketAddrResolver) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if resolver == nil {
			exception.Panicf("gate: %w: option WebSocketLocalAddrResolver can't be assigned to nil", core.ErrArgs)
		}
		options.WebSocketLocalAddrResolver = resolver
	}
}

// WebSocketRemoteAddrResolver 设置 WebSocket 客户端地址解析器，不得为 nil。
func (_GateOption) WebSocketRemoteAddrResolver(resolver WebSocketAddrResolver) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if resolver == nil {
			exception.Panicf("gate: %w: option WebSocketRemoteAddrResolver can't be assigned to nil", core.ErrArgs)
		}
		options.WebSocketRemoteAddrResolver = resolver
	}
}

// IOTimeout 设置单次网络 I/O 超时，必须不少于 100 毫秒。
func (_GateOption) IOTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 100*time.Millisecond {
			exception.Panicf("gate: %w: option IOTimeout must be >= 0.1 seconds", core.ErrArgs)
		}
		options.IOTimeout = d
	}
}

// IORetryTimes 设置网络 I/O 超时后的重试次数，必须大于等于 0。
func (_GateOption) IORetryTimes(times int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if times < 0 {
			exception.Panicf("gate: %w: option IORetryTimes must be >= 0", core.ErrArgs)
		}
		options.IORetryTimes = times
	}
}

// IOBufferCap 设置断线重连时保留的发送数据字节上限，必须不少于 1024。
func (_GateOption) IOBufferCap(cap int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if cap < 1024 {
			exception.Panicf("gate: %w: option IOBufferCap must be >= 1024 bytes", core.ErrArgs)
		}
		options.IOBufferCap = cap
	}
}

// MsgCreator 设置 GTP 消息构建器，不得为 nil。
func (_GateOption) MsgCreator(mc gtp.IMsgCreator) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if mc == nil {
			exception.Panicf("gate: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}

// AgreeClientEncryptionProposal 设置是否允许采用客户端建议的加密方案。
func (_GateOption) AgreeClientEncryptionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientEncryptionProposal = b
	}
}

// EncCipherSuite 设置服务端首选密码套件。
func (_GateOption) EncCipherSuite(cs gtp.CipherSuite) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncCipherSuite = cs
	}
}

// EncNonceStep 设置需要 nonce 的算法每次加解密后的增量。
func (_GateOption) EncNonceStep(v *big.Int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncNonceStep = v
	}
}

// EncECDHENamedCurve 设置 ECDHE 密钥交换使用的命名曲线。
func (_GateOption) EncECDHENamedCurve(nc gtp.NamedCurve) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncECDHENamedCurve = nc
	}
}

// EncSignatureAlgorithm 设置握手签名算法。
func (_GateOption) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

// EncSignaturePrivateKey 设置服务端握手签名私钥。
func (_GateOption) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

// EncVerifyClientSignature 设置是否验证客户端握手签名。
func (_GateOption) EncVerifyClientSignature(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifyClientSignature = b
	}
}

// EncVerifySignaturePublicKey 设置验证客户端握手签名使用的公钥。
func (_GateOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

// AgreeClientCompressionProposal 设置是否允许采用客户端建议的压缩方案。
func (_GateOption) AgreeClientCompressionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientCompressionProposal = b
	}
}

// Compression 设置服务端首选压缩算法。
func (_GateOption) Compression(c gtp.Compression) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.Compression = c
	}
}

// CompressionThreshold 设置启用压缩的字节阈值；小于等于 0 时禁用压缩。
func (_GateOption) CompressionThreshold(threshold int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.CompressionThreshold = threshold
	}
}

// MaxUncompressedSize 设置解压后负载字节上限，用于防御压缩炸弹。
func (_GateOption) MaxUncompressedSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.MaxUncompressedSize = size
	}
}

// MaxPacketSize 设置单个 GTP 包的字节上限。
func (_GateOption) MaxPacketSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.MaxPacketSize = size
	}
}

// AcceptTimeout 设置握手完成前的最长等待时间，必须不少于 300 毫秒。
func (_GateOption) AcceptTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("gate: %w: option AcceptTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.AcceptTimeout = d
	}
}

// Authenticator 设置客户端凭据校验器；nil 表示不额外鉴权。
func (_GateOption) Authenticator(auth Authenticator) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.Authenticator = auth
	}
}

// SessionInactiveTimeout 设置断线后等待连接迁移的时间，必须大于等于 0。
func (_GateOption) SessionInactiveTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 0 {
			exception.Panicf("gate: %w: option SessionInactiveTimeout must be >= 0 seconds", core.ErrArgs)
		}
		options.SessionInactiveTimeout = d
	}
}

// SessionWatcherInboxSize 设置每个会话观察器的收件箱容量，必须大于 0。
func (_GateOption) SessionWatcherInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionWatcherInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionWatcherInboxSize = size
	}
}

// SessionDataListenerInboxSize 设置每个会话数据监听器的收件箱容量，必须大于 0。
func (_GateOption) SessionDataListenerInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionDataListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionDataListenerInboxSize = size
	}
}

// SessionEventListenerInboxSize 设置每个会话事件监听器的收件箱容量，必须大于 0。
func (_GateOption) SessionEventListenerInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionEventListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionEventListenerInboxSize = size
	}
}

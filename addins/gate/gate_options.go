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
	WebSocketAddrResolver = generic.Func1[*websocket.Conn, net.Addr]                          // WebSocket的地址解析器
	Authenticator         = generic.Delegate5[IGate, net.Conn, string, string, []byte, error] // 鉴权客户端处理器（args: [gate, conn, userId, token, extensions], ret: [error]）
)

// GateOptions 网关所有选项
type GateOptions struct {
	TCPAddress                     string                 // TCP监听地址
	TCPNoDelay                     *bool                  // TCP的NoDelay选项，nil表示使用系统默认值
	TCPQuickAck                    *bool                  // TCP的QuickAck选项，nil表示使用系统默认值
	TCPRecvBuf                     *int                   // TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
	TCPSendBuf                     *int                   // TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
	TCPLinger                      *int                   // TCP的Linger选项，nil表示使用系统默认值
	TCPTLSConfig                   *tls.Config            // TCP的TLS配置，nil表示不使用TLS加密链路
	WebSocketURL                   *url.URL               // WebSocket监听地址
	WebSocketTLSConfig             *tls.Config            // WebSocket的TLS配置，nil表示不使用TLS加密链路
	WebSocketLocalAddrResolver     WebSocketAddrResolver  // WebSocket的本地地址解析器
	WebSocketRemoteAddrResolver    WebSocketAddrResolver  // WebSocket的对端地址解析器
	IOTimeout                      time.Duration          // 网络io超时时间
	IORetryTimes                   int                    // 网络io超时后的重试次数
	IOBufferCap                    int                    // 网络io缓存容量（字节）
	MsgCreator                     gtp.IMsgCreator        // 消息包解码器的消息构建器
	AgreeClientEncryptionProposal  bool                   // 是否同意使用客户端建议的加密方案
	EncCipherSuite                 gtp.CipherSuite        // 加密通信中的密码学套件
	EncNonceStep                   *big.Int               // 加密通信中，使用需要nonce的加密算法时，每次加解密自增值
	EncECDHENamedCurve             gtp.NamedCurve         // 加密通信中，在ECDHE交换秘钥时使用的曲线类型
	EncSignatureAlgorithm          gtp.SignatureAlgorithm // 加密通信中的签名算法
	EncSignaturePrivateKey         crypto.PrivateKey      // 加密通信中，签名用的私钥
	EncVerifyClientSignature       bool                   // 加密通信中，是否验证客户端签名
	EncVerifySignaturePublicKey    crypto.PublicKey       // 加密通信中，验证客户端签名用的公钥
	AgreeClientCompressionProposal bool                   // 是否同意使用客户端建议的压缩方案
	Compression                    gtp.Compression        // 通信中的压缩函数
	CompressionThreshold           int                    // 通信中启用压缩阀值（字节），<=0表示不开启
	AcceptTimeout                  time.Duration          // 接受连接超时时间
	Authenticator                  Authenticator          // 鉴权客户端处理器
	SessionInactiveTimeout         time.Duration          // 会话不活跃后的超时时间
	SessionWatcherInboxSize        int                    // 会话监听器inbox缓存大小
	SessionDataListenerInboxSize   int                    // 会话数据监听器inbox缓存大小
	SessionEventListenerInboxSize  int                    // 会话事件监听器inbox缓存大小
}

var With _GateOption

type _GateOption struct{}

// Default 默认选项
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
		With.AgreeClientEncryptionProposal(false)(options)
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
		With.AgreeClientCompressionProposal(false)(options)
		With.Compression(gtp.Compression_Brotli)(options)
		With.CompressionThreshold(64 * 1024)(options)
		With.AcceptTimeout(10 * time.Second)(options)
		With.Authenticator(nil)(options)
		With.SessionInactiveTimeout(time.Minute)(options)
		With.SessionWatcherInboxSize(256 * 1024)(options)
		With.SessionDataListenerInboxSize(128)(options)
		With.SessionEventListenerInboxSize(128)(options)
	}
}

// TCPAddress 设置TCP监听地址
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

// TCPNoDelay 设置TCP的NoDelay选项，nil表示使用系统默认值
func (_GateOption) TCPNoDelay(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPNoDelay = b
	}
}

// TCPQuickAck 设置TCP的QuickAck选项，nil表示使用系统默认值
func (_GateOption) TCPQuickAck(b *bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPQuickAck = b
	}
}

// TCPRecvBuf 设置TCP的RecvBuf大小（字节）选项，nil表示使用系统默认值
func (_GateOption) TCPRecvBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPRecvBuf = size
	}
}

// TCPSendBuf 设置TCP的SendBuf大小（字节）选项，nil表示使用系统默认值
func (_GateOption) TCPSendBuf(size *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPSendBuf = size
	}
}

// TCPLinger 设置TCP的Linger选项，nil表示使用系统默认值
func (_GateOption) TCPLinger(sec *int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPLinger = sec
	}
}

// TCPTLSConfig 设置TCP的TLS配置，nil表示不使用TLS加密链路
func (_GateOption) TCPTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.TCPTLSConfig = tlsConfig
	}
}

// WebSocketURL WebSocket监听地址
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

// WebSocketTLSConfig 设置WebSocket的TLS配置，nil表示不使用TLS加密链路
func (_GateOption) WebSocketTLSConfig(tlsConfig *tls.Config) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.WebSocketTLSConfig = tlsConfig
	}
}

// WebSocketLocalAddrResolver 设置WebSocket的本地地址解析器
func (_GateOption) WebSocketLocalAddrResolver(resolver WebSocketAddrResolver) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if resolver == nil {
			exception.Panicf("gate: %w: option WebSocketLocalAddrResolver can't be assigned to nil", core.ErrArgs)
		}
		options.WebSocketLocalAddrResolver = resolver
	}
}

// WebSocketRemoteAddrResolver 设置WebSocket的对端地址解析器
func (_GateOption) WebSocketRemoteAddrResolver(resolver WebSocketAddrResolver) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if resolver == nil {
			exception.Panicf("gate: %w: option WebSocketRemoteAddrResolver can't be assigned to nil", core.ErrArgs)
		}
		options.WebSocketRemoteAddrResolver = resolver
	}
}

// IOTimeout 设置网络io超时时间
func (_GateOption) IOTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 100*time.Millisecond {
			exception.Panicf("gate: %w: option IOTimeout must be >= 0.1 seconds", core.ErrArgs)
		}
		options.IOTimeout = d
	}
}

// IORetryTimes 设置网络io超时后的重试次数
func (_GateOption) IORetryTimes(times int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if times < 0 {
			exception.Panicf("gate: %w: option IORetryTimes must be >= 0", core.ErrArgs)
		}
		options.IORetryTimes = times
	}
}

// IOBufferCap 设置网络io缓存容量（字节）
func (_GateOption) IOBufferCap(cap int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if cap < 1024 {
			exception.Panicf("gate: %w: option IOBufferCap must be >= 1024 bytes", core.ErrArgs)
		}
		options.IOBufferCap = cap
	}
}

// MsgCreator 设置消息包解码器的消息构建器
func (_GateOption) MsgCreator(mc gtp.IMsgCreator) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if mc == nil {
			exception.Panicf("gate: %w: option MsgCreator can't be assigned to nil", core.ErrArgs)
		}
		options.MsgCreator = mc
	}
}

// AgreeClientEncryptionProposal 设置是否同意使用客户端建议的加密方案
func (_GateOption) AgreeClientEncryptionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientEncryptionProposal = b
	}
}

// EncCipherSuite 设置加密通信中的密码学套件
func (_GateOption) EncCipherSuite(cs gtp.CipherSuite) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncCipherSuite = cs
	}
}

// EncNonceStep 设置加密通信中，使用需要nonce的加密算法时，每次加解密自增值
func (_GateOption) EncNonceStep(v *big.Int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncNonceStep = v
	}
}

// EncECDHENamedCurve 设置加密通信中，在ECDHE交换秘钥时使用的曲线类型
func (_GateOption) EncECDHENamedCurve(nc gtp.NamedCurve) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncECDHENamedCurve = nc
	}
}

// EncSignatureAlgorithm 设置加密通信中的签名算法
func (_GateOption) EncSignatureAlgorithm(sa gtp.SignatureAlgorithm) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignatureAlgorithm = sa
	}
}

// EncSignaturePrivateKey 设置加密通信中，签名用的私钥
func (_GateOption) EncSignaturePrivateKey(priv crypto.PrivateKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncSignaturePrivateKey = priv
	}
}

// EncVerifyClientSignature 设置加密通信中，是否验证客户端签名
func (_GateOption) EncVerifyClientSignature(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifyClientSignature = b
	}
}

// EncVerifySignaturePublicKey 设置加密通信中，验证客户端签名用的公钥
func (_GateOption) EncVerifySignaturePublicKey(pub crypto.PublicKey) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.EncVerifySignaturePublicKey = pub
	}
}

// AgreeClientCompressionProposal 设置是否同意使用客户端建议的压缩方案
func (_GateOption) AgreeClientCompressionProposal(b bool) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.AgreeClientCompressionProposal = b
	}
}

// Compression 设置通信中的压缩函数
func (_GateOption) Compression(c gtp.Compression) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.Compression = c
	}
}

// CompressionThreshold 设置通信中启用压缩阀值（字节），<=0表示不开启
func (_GateOption) CompressionThreshold(threshold int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.CompressionThreshold = threshold
	}
}

// AcceptTimeout 设置接受连接超时时间
func (_GateOption) AcceptTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 300*time.Millisecond {
			exception.Panicf("gate: %w: option AcceptTimeout must be >= 0.3 seconds", core.ErrArgs)
		}
		options.AcceptTimeout = d
	}
}

// Authenticator 设置鉴权客户端处理器
func (_GateOption) Authenticator(auth Authenticator) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		options.Authenticator = auth
	}
}

// SessionInactiveTimeout 设置会话不活跃后的超时时间
func (_GateOption) SessionInactiveTimeout(d time.Duration) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if d < 0 {
			exception.Panicf("gate: %w: option SessionInactiveTimeout must be >= 0 seconds", core.ErrArgs)
		}
		options.SessionInactiveTimeout = d
	}
}

// SessionWatcherInboxSize 设置会话监听器inbox缓存大小
func (_GateOption) SessionWatcherInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionWatcherInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionWatcherInboxSize = size
	}
}

// SessionDataListenerInboxSize 设置会话数据监听器inbox缓存大小
func (_GateOption) SessionDataListenerInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionDataListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionDataListenerInboxSize = size
	}
}

// SessionEventListenerInboxSize 设置会话事件监听器inbox缓存大小
func (_GateOption) SessionEventListenerInboxSize(size int) option.Setting[GateOptions] {
	return func(options *GateOptions) {
		if size <= 0 {
			exception.Panicf("gate: %w: option SessionEventListenerInboxSize must be > 0", core.ErrArgs)
		}
		options.SessionEventListenerInboxSize = size
	}
}

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

package rpcli

import (
	"context"
	"crypto"
	"crypto/tls"
	"fmt"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/gate/cli"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gtp"
	"go.uber.org/zap"
)

// BuildRPCli 创建采用默认消息构建器、三次时钟采样并启用调用路径压缩的 RPC 客户端构建器。
func BuildRPCli() *RPCliCreator {
	return &RPCliCreator{
		timeSamples:    3,
		msgCreator:     gap.DefaultMsgCreator(),
		reduceCallPath: true,
	}
}

// RPCliCreator 分步配置并连接 RPC 客户端。
type RPCliCreator struct {
	settings       []option.Setting[cli.ClientOptions]
	timeSamples    int
	msgCreator     gap.IMsgCreator
	reduceCallPath bool
	scripts        generic.SliceMap[string, IScript]
}

// SetNetProtocol 设置底层连接协议。
func (ctor *RPCliCreator) SetNetProtocol(p cli.NetProtocol) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.NetProtocol(p))
	return ctor
}

// SetTCPNoDelay 设置 TCP_NODELAY。
func (ctor *RPCliCreator) SetTCPNoDelay(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPNoDelay(&b))
	return ctor
}

// SetTCPQuickAck 设置 TCP_QUICKACK。
func (ctor *RPCliCreator) SetTCPQuickAck(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPQuickAck(&b))
	return ctor
}

// SetTCPRecvBuf 设置 TCP 接收缓冲区字节数。
func (ctor *RPCliCreator) SetTCPRecvBuf(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPRecvBuf(&size))
	return ctor
}

// SetTCPSendBuf 设置 TCP 发送缓冲区字节数。
func (ctor *RPCliCreator) SetTCPSendBuf(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPSendBuf(&size))
	return ctor
}

// SetTCPLinger 设置 TCP 关闭等待秒数。
func (ctor *RPCliCreator) SetTCPLinger(sec int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPLinger(&sec))
	return ctor
}

// SetWebSocketOrigin 设置 WebSocket Origin。
func (ctor *RPCliCreator) SetWebSocketOrigin(origin string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.WebSocketOrigin(origin))
	return ctor
}

// SetTLSConfig 设置底层 TLS 配置；nil 表示不额外启用 TLS。
func (ctor *RPCliCreator) SetTLSConfig(tlsConfig *tls.Config) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TLSConfig(tlsConfig))
	return ctor
}

// SetIOTimeout 设置单次网络 I/O 超时，必须不少于 100 毫秒。
func (ctor *RPCliCreator) SetIOTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOTimeout(d))
	return ctor
}

// SetIORetryTimes 设置网络 I/O 超时后的重试次数，必须大于等于 0。
func (ctor *RPCliCreator) SetIORetryTimes(times int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IORetryTimes(times))
	return ctor
}

// SetIOBufferCap 设置断线重连时保留的发送数据字节上限，必须不少于 1024。
func (ctor *RPCliCreator) SetIOBufferCap(cap int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOBufferCap(cap))
	return ctor
}

// SetGTPDecoderMsgCreator 设置 GTP 消息构建器，不得为 nil。
func (ctor *RPCliCreator) SetGTPDecoderMsgCreator(mc gtp.IMsgCreator) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.MsgCreator(mc))
	return ctor
}

// SetGTPEncCipherSuite 设置客户端向服务端提出的密码套件。
func (ctor *RPCliCreator) SetGTPEncCipherSuite(cs gtp.CipherSuite) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncCipherSuite(cs))
	return ctor
}

// SetGTPEncSignatureAlgorithm 设置客户端握手签名算法。
func (ctor *RPCliCreator) SetGTPEncSignatureAlgorithm(sa gtp.SignatureAlgorithm) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignatureAlgorithm(sa))
	return ctor
}

// SetGTPEncSignaturePrivateKey 设置客户端握手签名私钥。
func (ctor *RPCliCreator) SetGTPEncSignaturePrivateKey(priv crypto.PrivateKey) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignaturePrivateKey(priv))
	return ctor
}

// SetGTPEncVerifyServerSignature 设置是否验证服务端握手签名。
func (ctor *RPCliCreator) SetGTPEncVerifyServerSignature(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifyServerSignature(b))
	return ctor
}

// SetGTPEncVerifySignaturePublicKey 设置验证服务端握手签名使用的公钥。
func (ctor *RPCliCreator) SetGTPEncVerifySignaturePublicKey(pub crypto.PublicKey) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifySignaturePublicKey(pub))
	return ctor
}

// SetGTPCompression 设置客户端向服务端提出的压缩算法。
func (ctor *RPCliCreator) SetGTPCompression(c gtp.Compression) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.Compression(c))
	return ctor
}

// SetGTPCompressedSize 设置启用 GTP 压缩的字节阈值；小于等于 0 时禁用压缩。
func (ctor *RPCliCreator) SetGTPCompressedSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.CompressionThreshold(size))
	return ctor
}

// SetGTPMaxUncompressedSize 设置解压后负载字节上限。
func (ctor *RPCliCreator) SetGTPMaxUncompressedSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.MaxUncompressedSize(size))
	return ctor
}

// SetGTPAutoReconnect 设置连接失活后是否自动迁移会话连接。
func (ctor *RPCliCreator) SetGTPAutoReconnect(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnect(b))
	return ctor
}

// SetGTPAutoReconnectInterval 设置相邻重连尝试的间隔，必须大于等于 0。
func (ctor *RPCliCreator) SetGTPAutoReconnectInterval(dur time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectInterval(dur))
	return ctor
}

// SetGTPAutoReconnectRetryTimes 设置自动重连尝试次数；小于等于 0 时无限重试。
func (ctor *RPCliCreator) SetGTPAutoReconnectRetryTimes(times int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectRetryTimes(times))
	return ctor
}

// SetGTPInactiveTimeout 设置未启用自动重连时的失活等待时间，必须大于等于 0。
func (ctor *RPCliCreator) SetGTPInactiveTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.InactiveTimeout(d))
	return ctor
}

// SetGTPTimeSamples 设置连接后用于选择最低 RTT 时钟样本的采样次数，最少为 3。
func (ctor *RPCliCreator) SetGTPTimeSamples(n int) *RPCliCreator {
	if n < 3 {
		exception.Panicf("rpcli: %w: option GTPTimeSamples can't be set to a value less than 3", core.ErrArgs)
	}
	ctor.timeSamples = n
	return ctor
}

// SetGAPDecoderMsgCreator 设置 GAP 消息构建器，不得为 nil。
func (ctor *RPCliCreator) SetGAPDecoderMsgCreator(mc gap.IMsgCreator) *RPCliCreator {
	ctor.msgCreator = mc
	return ctor
}

// SetFutureTimeout 设置关联请求 Future 的默认超时，必须不少于 300 毫秒。
func (ctor *RPCliCreator) SetFutureTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.FutureTimeout(d))
	return ctor
}

// SetAuthUserID 设置握手提交的用户 ID。
func (ctor *RPCliCreator) SetAuthUserID(userID string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthUserID(userID))
	return ctor
}

// SetAuthToken 设置握手提交的鉴权令牌。
func (ctor *RPCliCreator) SetAuthToken(token string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthToken(token))
	return ctor
}

// SetAuthExtensions 设置握手提交的扩展数据；切片不会复制。
func (ctor *RPCliCreator) SetAuthExtensions(extensions []byte) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthExtensions(extensions))
	return ctor
}

// SetPanicHandling 设置监听器是否自动恢复 panic 以及错误报告通道。
func (ctor *RPCliCreator) SetPanicHandling(autoRecover bool, reportError chan error) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.PanicHandling(autoRecover, reportError))
	return ctor
}

// SetDataListenerInboxSize 设置每个数据监听器的收件箱容量，必须大于 0。
func (ctor *RPCliCreator) SetDataListenerInboxSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.DataListenerInboxSize(size))
	return ctor
}

// SetEventListenerInboxSize 设置每个事件监听器的收件箱容量，必须大于 0。
func (ctor *RPCliCreator) SetEventListenerInboxSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EventListenerInboxSize(size))
	return ctor
}

// SetReduceCallPath 设置是否用进程内缓存索引压缩脚本名和方法名。
func (ctor *RPCliCreator) SetReduceCallPath(b bool) *RPCliCreator {
	ctor.reduceCallPath = b
	return ctor
}

// SetScripts 注册客户端可被远端调用的脚本；名称不得重复，脚本不得为 nil。
func (ctor *RPCliCreator) SetScripts(scripts map[string]IScript) *RPCliCreator {
	for name, script := range scripts {
		if script == nil {
			exception.Panicf("rpcli: %w: script %q can't be nil", core.ErrArgs, name)
		}
		if ctor.scripts.Exist(name) {
			exception.Panicf("rpcli: %w: script %q has been registered", core.ErrArgs, name)
		}
		ctor.scripts.Add(name, script)
	}
	return ctor
}

// SetLogger 设置客户端日志器；nil 表示使用不输出日志的实现。
func (ctor *RPCliCreator) SetLogger(logger *zap.Logger) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.Logger(logger))
	return ctor
}

// Connect 连接端点、完成时钟采样并启动 GAP 消息监听。
func (ctor *RPCliCreator) Connect(ctx context.Context, endpoint string) (*RPCli, error) {
	client, err := cli.Connect(ctx, endpoint, ctor.settings...)
	if err != nil {
		return nil, err
	}

	var remoteClock *cli.TimeSample

	for range ctor.timeSamples {
		resp := client.ProbeTime().Wait(ctx)
		if !resp.OK() {
			client.Close(resp.Error)
			return nil, resp.Error
		}

		current, ok := resp.Value.(*cli.TimeSample)
		if !ok {
			err := fmt.Errorf("rpcli: unexpected probe time response type, expected *cli.TimeSample, got %T", resp.Value)
			client.Close(err)
			return nil, err
		}

		if remoteClock != nil {
			if current.RTT() < remoteClock.RTT() {
				remoteClock = current
			}
		} else {
			remoteClock = current
		}
	}

	rpcli := &RPCli{
		Client:         client,
		encoder:        codec.NewEncoder(),
		decoder:        codec.NewDecoder(ctor.msgCreator),
		remoteClock:    *remoteClock,
		reduceCallPath: ctor.reduceCallPath,
	}

	ctor.scripts.Each(func(name string, script IScript) {
		script.init(rpcli, name, script)
		cacheCallPath(name, script.Reflected().Type())

		rpcli.scripts.Add(name, script)
	})

	rpcli.DataIO().Listen(context.Background(), generic.CastDelegateVoid1(rpcli.handleData))

	return rpcli, nil
}

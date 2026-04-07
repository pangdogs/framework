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

// BuildRPCli 创建RPC客户端
func BuildRPCli() *RPCliCreator {
	return &RPCliCreator{
		rttSampling:    3,
		msgCreator:     gap.DefaultMsgCreator(),
		reduceCallPath: true,
	}
}

// RPCliCreator RPC客户端构建器
type RPCliCreator struct {
	settings       []option.Setting[cli.ClientOptions]
	rttSampling    int
	msgCreator     gap.IMsgCreator
	reduceCallPath bool
	scripts        generic.SliceMap[string, IScript]
}

func (ctor *RPCliCreator) SetNetProtocol(p cli.NetProtocol) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.NetProtocol(p))
	return ctor
}

func (ctor *RPCliCreator) SetTCPNoDelay(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPNoDelay(&b))
	return ctor
}

func (ctor *RPCliCreator) SetTCPQuickAck(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPQuickAck(&b))
	return ctor
}

func (ctor *RPCliCreator) SetTCPRecvBuf(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPRecvBuf(&size))
	return ctor
}

func (ctor *RPCliCreator) SetTCPSendBuf(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPSendBuf(&size))
	return ctor
}

func (ctor *RPCliCreator) SetTCPLinger(sec int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPLinger(&sec))
	return ctor
}

func (ctor *RPCliCreator) SetWebSocketOrigin(origin string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.WebSocketOrigin(origin))
	return ctor
}

func (ctor *RPCliCreator) SetTLSConfig(tlsConfig *tls.Config) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TLSConfig(tlsConfig))
	return ctor
}

func (ctor *RPCliCreator) SetIOTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOTimeout(d))
	return ctor
}

func (ctor *RPCliCreator) SetIORetryTimes(times int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IORetryTimes(times))
	return ctor
}

func (ctor *RPCliCreator) SetIOBufferCap(cap int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOBufferCap(cap))
	return ctor
}

func (ctor *RPCliCreator) SetGTPDecoderMsgCreator(mc gtp.IMsgCreator) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.MsgCreator(mc))
	return ctor
}

func (ctor *RPCliCreator) SetGTPEncCipherSuite(cs gtp.CipherSuite) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncCipherSuite(cs))
	return ctor
}

func (ctor *RPCliCreator) SetGTPEncSignatureAlgorithm(sa gtp.SignatureAlgorithm) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignatureAlgorithm(sa))
	return ctor
}

func (ctor *RPCliCreator) SetGTPEncSignaturePrivateKey(priv crypto.PrivateKey) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignaturePrivateKey(priv))
	return ctor
}

func (ctor *RPCliCreator) SetGTPEncVerifyServerSignature(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifyServerSignature(b))
	return ctor
}

func (ctor *RPCliCreator) SetGTPEncVerifySignaturePublicKey(pub crypto.PublicKey) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifySignaturePublicKey(pub))
	return ctor
}

func (ctor *RPCliCreator) SetGTPCompression(c gtp.Compression) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.Compression(c))
	return ctor
}

func (ctor *RPCliCreator) SetGTPCompressedSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.CompressionThreshold(size))
	return ctor
}

func (ctor *RPCliCreator) SetGTPMaxUncompressedSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.MaxUncompressedSize(size))
	return ctor
}

func (ctor *RPCliCreator) SetGTPAutoReconnect(b bool) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnect(b))
	return ctor
}

func (ctor *RPCliCreator) SetGTPAutoReconnectInterval(dur time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectInterval(dur))
	return ctor
}

func (ctor *RPCliCreator) SetGTPAutoReconnectRetryTimes(times int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectRetryTimes(times))
	return ctor
}

func (ctor *RPCliCreator) SetGTPInactiveTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.InactiveTimeout(d))
	return ctor
}

func (ctor *RPCliCreator) SetGTPRTTSampling(n int) *RPCliCreator {
	if n < 3 {
		exception.Panicf("rpcli: %w: option GTPRTTSampling can't be set to a value less than 3", core.ErrArgs)
	}
	ctor.rttSampling = n
	return ctor
}

func (ctor *RPCliCreator) SetGAPDecoderMsgCreator(mc gap.IMsgCreator) *RPCliCreator {
	ctor.msgCreator = mc
	return ctor
}

func (ctor *RPCliCreator) SetFutureTimeout(d time.Duration) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.FutureTimeout(d))
	return ctor
}

func (ctor *RPCliCreator) SetAuthUserId(userId string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthUserId(userId))
	return ctor
}

func (ctor *RPCliCreator) SetAuthToken(token string) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthToken(token))
	return ctor
}

func (ctor *RPCliCreator) SetAuthExtensions(extensions []byte) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthExtensions(extensions))
	return ctor
}

func (ctor *RPCliCreator) SetPanicHandling(autoRecover bool, reportError chan error) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.PanicHandling(autoRecover, reportError))
	return ctor
}

func (ctor *RPCliCreator) SetDataListenerInboxSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.DataListenerInboxSize(size))
	return ctor
}

func (ctor *RPCliCreator) SetEventListenerInboxSize(size int) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EventListenerInboxSize(size))
	return ctor
}

func (ctor *RPCliCreator) SetReduceCallPath(b bool) *RPCliCreator {
	ctor.reduceCallPath = b
	return ctor
}

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

func (ctor *RPCliCreator) SetLogger(logger *zap.Logger) *RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.Logger(logger))
	return ctor
}

func (ctor *RPCliCreator) Connect(ctx context.Context, endpoint string) (*RPCli, error) {
	client, err := cli.Connect(ctx, endpoint, ctor.settings...)
	if err != nil {
		return nil, err
	}

	var remoteTime *cli.ResponseTime

	for range ctor.rttSampling {
		respTime := client.RequestTime().Wait(ctx)
		if !respTime.OK() {
			client.Close(respTime.Error)
			return nil, respTime.Error
		}

		current, ok := respTime.Value.(*cli.ResponseTime)
		if !ok {
			err := fmt.Errorf("rpcli: unexpected response time type %T", respTime.Value)
			client.Close(err)
			return nil, err
		}

		if remoteTime != nil {
			if current.RTT() < remoteTime.RTT() {
				remoteTime = current
			}
		} else {
			remoteTime = current
		}
	}

	rpcli := &RPCli{
		Client:         client,
		encoder:        codec.NewEncoder(),
		decoder:        codec.NewDecoder(ctor.msgCreator),
		remoteTime:     *remoteTime,
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

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
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/gate/cli"
	"git.golaxy.org/framework/net/gap"
	"git.golaxy.org/framework/net/gap/codec"
	"git.golaxy.org/framework/net/gtp"
	"go.uber.org/zap"
	"time"
)

// CreateRPCli 创建RPC客户端
func CreateRPCli() RPCliCreator {
	return RPCliCreator{
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
	mainProc       IProcedure
}

func (ctor RPCliCreator) NetProtocol(p cli.NetProtocol) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.NetProtocol(p))
	return ctor
}

func (ctor RPCliCreator) TCPNoDelay(b bool) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPNoDelay(&b))
	return ctor
}

func (ctor RPCliCreator) TCPQuickAck(b bool) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPQuickAck(&b))
	return ctor
}

func (ctor RPCliCreator) TCPRecvBuf(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPRecvBuf(&size))
	return ctor
}

func (ctor RPCliCreator) TCPSendBuf(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPSendBuf(&size))
	return ctor
}

func (ctor RPCliCreator) TCPLinger(sec int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TCPLinger(&sec))
	return ctor
}

func (ctor RPCliCreator) WebSocketOrigin(origin string) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.WebSocketOrigin(origin))
	return ctor
}

func (ctor RPCliCreator) TLSConfig(tlsConfig tls.Config) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.TLSConfig(&tlsConfig))
	return ctor
}

func (ctor RPCliCreator) IOTimeout(d time.Duration) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOTimeout(d))
	return ctor
}

func (ctor RPCliCreator) IORetryTimes(times int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IORetryTimes(times))
	return ctor
}

func (ctor RPCliCreator) IOBufferCap(cap int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.IOBufferCap(cap))
	return ctor
}

func (ctor RPCliCreator) GTPDecoderMsgCreator(mc gtp.IMsgCreator) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.DecoderMsgCreator(mc))
	return ctor
}

func (ctor RPCliCreator) GTPEncCipherSuite(cs gtp.CipherSuite) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncCipherSuite(cs))
	return ctor
}

func (ctor RPCliCreator) GTPEncSignatureAlgorithm(sa gtp.SignatureAlgorithm) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignatureAlgorithm(sa))
	return ctor
}

func (ctor RPCliCreator) GTPEncSignaturePrivateKey(priv crypto.PrivateKey) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncSignaturePrivateKey(priv))
	return ctor
}

func (ctor RPCliCreator) GTPEncVerifyServerSignature(b bool) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifyServerSignature(b))
	return ctor
}

func (ctor RPCliCreator) GTPEncVerifySignaturePublicKey(pub crypto.PublicKey) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.EncVerifySignaturePublicKey(pub))
	return ctor
}

func (ctor RPCliCreator) GTPCompression(c gtp.Compression) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.Compression(c))
	return ctor
}

func (ctor RPCliCreator) GTPCompressedSize(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.CompressedSize(size))
	return ctor
}

func (ctor RPCliCreator) GTPAutoReconnect(b bool) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnect(b))
	return ctor
}

func (ctor RPCliCreator) GTPAutoReconnectInterval(dur time.Duration) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectInterval(dur))
	return ctor
}

func (ctor RPCliCreator) GTPAutoReconnectRetryTimes(times int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AutoReconnectRetryTimes(times))
	return ctor
}

func (ctor RPCliCreator) GTPInactiveTimeout(d time.Duration) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.InactiveTimeout(d))
	return ctor
}

func (ctor RPCliCreator) GTPSendDataChanSize(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.SendDataChanSize(size))
	return ctor
}

func (ctor RPCliCreator) GTPRecvDataChanSize(size int, recyclable bool) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.RecvDataChanSize(size, recyclable))
	return ctor
}

func (ctor RPCliCreator) GTPSendEventChanSize(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.SendEventChanSize(size))
	return ctor
}

func (ctor RPCliCreator) GTPRecvEventChanSize(size int) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.RecvEventChanSize(size))
	return ctor
}

func (ctor RPCliCreator) GTPRecvDataHandler(handler cli.RecvDataHandler) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.RecvDataHandler(handler))
	return ctor
}

func (ctor RPCliCreator) GTPRecvEventHandler(handler cli.RecvEventHandler) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.RecvEventHandler(handler))
	return ctor
}

func (ctor RPCliCreator) GTPRTTSampling(n int) RPCliCreator {
	if n < 3 {
		exception.Panicf("%w: option GTPRTTSampling can't be set to a value less than 3", core.ErrArgs)
	}
	ctor.rttSampling = n
	return ctor
}

func (ctor RPCliCreator) GAPDecoderMsgCreator(mc gap.IMsgCreator) RPCliCreator {
	ctor.msgCreator = mc
	return ctor
}

func (ctor RPCliCreator) FutureTimeout(d time.Duration) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.FutureTimeout(d))
	return ctor
}

func (ctor RPCliCreator) AuthUserId(userId string) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthUserId(userId))
	return ctor
}

func (ctor RPCliCreator) AuthToken(token string) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthToken(token))
	return ctor
}

func (ctor RPCliCreator) AuthExtensions(extensions []byte) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.AuthExtensions(extensions))
	return ctor
}

func (ctor RPCliCreator) ReduceCallPath(b bool) RPCliCreator {
	ctor.reduceCallPath = b
	return ctor
}

func (ctor RPCliCreator) MainProcedure(proc any) RPCliCreator {
	_proc, ok := proc.(IProcedure)
	if !ok {
		exception.Panicf("%w: incorrect proc type", core.ErrArgs)
	}
	ctor.mainProc = _proc
	return ctor
}

func (ctor RPCliCreator) ZapLogger(logger *zap.Logger) RPCliCreator {
	ctor.settings = append(ctor.settings, cli.With.ZapLogger(logger))
	return ctor
}

func (ctor RPCliCreator) Connect(ctx context.Context, endpoint string) (*RPCli, error) {
	client, err := cli.Connect(ctx, endpoint, ctor.settings...)
	if err != nil {
		return nil, err
	}

	var remoteTime *cli.ResponseTime

	for range ctor.rttSampling {
		respTime := <-client.RequestTime(ctx)
		if !respTime.OK() {
			return nil, respTime.Error
		}

		if remoteTime != nil {
			if respTime.Value.RTT() < remoteTime.RTT() {
				remoteTime = respTime.Value
			}
		} else {
			remoteTime = respTime.Value
		}
	}

	rpcli := &RPCli{
		Client:         client,
		encoder:        codec.MakeEncoder(),
		decoder:        codec.MakeDecoder(ctor.msgCreator),
		remoteTime:     *remoteTime,
		reduceCallPath: ctor.reduceCallPath,
	}

	if ctor.mainProc != nil {
		rpcli.AddProcedure(Main, ctor.mainProc)
	}

	rpcli.WatchData(context.Background(), generic.CastDelegate1(rpcli.handleRecvData))

	return rpcli, nil
}

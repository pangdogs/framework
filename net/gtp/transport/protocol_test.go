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

package transport_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/transport"
)

func Test_Protocol(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:7000")
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var wg sync.WaitGroup

	// 服务端
	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				conn.Close()
			}()

			transceiver := &transport.Transceiver{
				Conn:         conn,
				Encoder:      codec.NewEncoder(),
				Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
				Timeout:      5 * time.Second,
				Synchronizer: transport.NewUnsequencedSynchronizer(),
			}

			handshake := &transport.HandshakeProtocol{
				Transceiver: transceiver,
			}

			err = handshake.ServerHello(context.Background(), func(e transport.Event[*gtp.MsgHello]) (transport.Event[*gtp.MsgHello], error) {
				log.Println("server <= hello")
				return transport.Event[*gtp.MsgHello]{Flags: gtp.Flags(gtp.Flag_HelloDone), Msg: &gtp.MsgHello{}}, nil
			})
			if err != nil {
				log.Panic(err)
			}

			err = handshake.ServerFinished(context.Background(), transport.Event[*gtp.MsgFinished]{
				Msg: &gtp.MsgFinished{
					SendSeq: transceiver.Synchronizer.SendSeq(),
					RecvSeq: transceiver.Synchronizer.RecvSeq(),
				},
			})
			if err != nil {
				log.Panic(err)
			}

			ctrl := &transport.CtrlProtocol{
				Transceiver: transceiver,
				HeartbeatHandler: generic.CastDelegateVoid1(func(e transport.Event[*gtp.MsgHeartbeat]) {
					text := "ping"
					if e.Flags.Is(gtp.Flag_Pong) {
						text = "pong"
					}
					log.Printf("server <= seq:%d ack:%d %s", e.Seq, e.Ack, text)
				}),
			}

			trans := &transport.TransProtocol{
				Transceiver: transceiver,
				PayloadHandler: generic.CastDelegateVoid1(func(e transport.Event[*gtp.MsgPayload]) {
					log.Printf("server <= seq:%d ack:%d data:%q", e.Seq, e.Ack, string(e.Msg.Data))
				}),
			}

			dispatcher := transport.EventDispatcher{
				Transceiver:  transceiver,
				EventHandler: generic.CastDelegateVoid1(ctrl.HandleEvent, trans.HandleEvent),
			}

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}

					ds := fmt.Sprintf("hello world, %d", rand.Uint64())

					err := trans.SendData([]byte(ds))
					if err != nil {
						log.Panic(err)
					}

					log.Println("server =>", ds)

					time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
				}
			}()

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}

					err := ctrl.SendPing()
					if err != nil {
						log.Panic(err)
					}

					log.Println("server => ping")

					time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
				}
			}()

			defer dispatcher.Transceiver.Dispose()

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				err := dispatcher.Dispatch(ctx)
				if err != nil {
					log.Println("server <= err:", err)
				}
			}
		}()
	}()

	// 客户端
	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", "127.0.0.1:7000")
		if err != nil {
			log.Panic(err)
		}
		defer conn.Close()

		transceiver := &transport.Transceiver{
			Conn:         conn,
			Encoder:      codec.NewEncoder(),
			Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
			Timeout:      5 * time.Second,
			Synchronizer: transport.NewUnsequencedSynchronizer(),
		}

		handshake := &transport.HandshakeProtocol{
			Transceiver: transceiver,
		}

		err = handshake.ClientHello(context.Background(), transport.Event[*gtp.MsgHello]{Msg: &gtp.MsgHello{}}, func(e transport.Event[*gtp.MsgHello]) error {
			log.Println("client <= hello")
			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		err = handshake.ClientFinished(context.Background(), func(e transport.Event[*gtp.MsgFinished]) error {
			log.Println("client <= finished", e.Msg.SendSeq, e.Msg.RecvSeq)
			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		ctrl := &transport.CtrlProtocol{
			Transceiver: transceiver,
			HeartbeatHandler: generic.CastDelegateVoid1(func(e transport.Event[*gtp.MsgHeartbeat]) {
				text := "ping"
				if e.Flags.Is(gtp.Flag_Pong) {
					text = "pong"
				}
				log.Printf("client <= seq:%d ack:%d %s", e.Seq, e.Ack, text)
			}),
			SyncTimeHandler: generic.CastDelegateVoid1(func(e transport.Event[*gtp.MsgSyncTime]) {
				log.Printf("client <= response time %d %d", e.Msg.LocalTime, e.Msg.RemoteTime)
			}),
		}

		trans := &transport.TransProtocol{
			Transceiver: transceiver,
			PayloadHandler: generic.CastDelegateVoid1(func(e transport.Event[*gtp.MsgPayload]) {
				log.Printf("client <= seq:%d ack:%d data:%q", e.Seq, e.Ack, string(e.Msg.Data))
			}),
		}

		dispatcher := transport.EventDispatcher{
			Transceiver:  transceiver,
			EventHandler: generic.CastDelegateVoid1(ctrl.HandleEvent, trans.HandleEvent),
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				err := ctrl.RequestTime(0)
				if err != nil {
					log.Panic(err)
				}

				log.Println("client => request time")

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				ds := fmt.Sprintf("hello world, %d", rand.Uint64())

				err := trans.SendData([]byte(ds))
				if err != nil {
					log.Panic(err)
				}

				log.Println("client =>", ds)

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				err := ctrl.SendPing()
				if err != nil {
					log.Panic(err)
				}

				log.Println("client => ping")

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		defer dispatcher.Transceiver.Dispose()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			err := dispatcher.Dispatch(ctx)
			if err != nil {
				log.Println("client <= err:", err)
			}
		}
	}()

	wg.Wait()
}

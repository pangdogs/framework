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

package transport

import (
	"context"
	"fmt"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func TestProtocol(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:7000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	var wg sync.WaitGroup

	// 服务端
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					conn.Close()
				}()

				transceiver := &Transceiver{
					Conn:         conn,
					Encoder:      codec.NewEncoder(),
					Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
					Synchronizer: NewUnsequencedSynchronizer(),
				}

				handshake := &HandshakeProtocol{
					Transceiver: transceiver,
				}

				err = handshake.ServerHello(context.Background(), func(e Event[gtp.MsgHello]) (Event[gtp.MsgHello], error) {
					fmt.Println(time.Now().Format(time.RFC3339Nano), "server <= hello")
					return Event[gtp.MsgHello]{Flags: gtp.Flags(gtp.Flag_HelloDone)}, nil
				})
				if err != nil {
					panic(err)
				}

				err = handshake.ServerFinished(context.Background(), Event[gtp.MsgFinished]{
					Msg: gtp.MsgFinished{
						SendSeq: transceiver.Synchronizer.SendSeq(),
						RecvSeq: transceiver.Synchronizer.RecvSeq(),
					},
				})
				if err != nil {
					panic(err)
				}

				ctrl := &CtrlProtocol{
					Transceiver: transceiver,
					HeartbeatHandler: generic.CastDelegate1(func(e Event[gtp.MsgHeartbeat]) error {
						text := "ping"
						if e.Flags.Is(gtp.Flag_Pong) {
							text = "pong"
						}
						fmt.Printf("%s server <= seq:%d ack:%d %s\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, text)
						return nil
					}),
				}

				trans := &TransProtocol{
					Transceiver: transceiver,
					PayloadHandler: generic.CastDelegate1(func(e Event[gtp.MsgPayload]) error {
						fmt.Printf("%s server <= seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, string(e.Msg.Data))
						return nil
					}),
				}

				dispatcher := EventDispatcher{
					Transceiver:  transceiver,
					EventHandler: generic.CastDelegate1(ctrl.HandleRecvEvent, trans.HandleRecvEvent),
				}

				go func() {
					for {
						ds := fmt.Sprintf("hello world, %d", rand.Uint64())

						err := trans.SendData([]byte(ds))
						if err != nil {
							panic(err)
						}

						fmt.Println(time.Now().Format(time.RFC3339Nano), "server =>", ds)

						time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
					}
				}()

				go func() {
					for {
						for {
							err := ctrl.SendPing()
							if err != nil {
								panic(err)
							}

							fmt.Println(time.Now().Format(time.RFC3339Nano), "server => ping")

							time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
						}
					}
				}()

				dispatcher.Run(context.Background(), generic.CastDelegateVoid1(func(err error) {
					fmt.Println(time.Now().Format(time.RFC3339Nano), "server <= err:", err)
				}))
			}()
		}
	}()

	// 客户端
	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", "127.0.0.1:7000")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		transceiver := &Transceiver{
			Conn:         conn,
			Encoder:      codec.NewEncoder(),
			Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
			Synchronizer: NewUnsequencedSynchronizer(),
		}

		handshake := &HandshakeProtocol{
			Transceiver: transceiver,
		}

		err = handshake.ClientHello(context.Background(), Event[gtp.MsgHello]{}, func(e Event[gtp.MsgHello]) error {
			fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= hello")
			return nil
		})
		if err != nil {
			panic(err)
		}

		err = handshake.ClientFinished(context.Background(), func(e Event[gtp.MsgFinished]) error {
			fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= finished", e.Msg.SendSeq, e.Msg.RecvSeq)
			return nil
		})
		if err != nil {
			panic(err)
		}

		ctrl := &CtrlProtocol{
			Transceiver: transceiver,
			HeartbeatHandler: generic.CastDelegate1(func(e Event[gtp.MsgHeartbeat]) error {
				text := "ping"
				if e.Flags.Is(gtp.Flag_Pong) {
					text = "pong"
				}
				fmt.Printf("%s client <= seq:%d ack:%d %s\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, text)
				return nil
			}),
			SyncTimeHandler: generic.CastDelegate1(func(e Event[gtp.MsgSyncTime]) error {
				fmt.Printf("%s client <= response time %d %d\n", time.Now().Format(time.RFC3339Nano), e.Msg.LocalTime, e.Msg.RemoteTime)
				return nil
			}),
		}

		trans := &TransProtocol{
			Transceiver: transceiver,
			PayloadHandler: generic.CastDelegate1(func(e Event[gtp.MsgPayload]) error {
				fmt.Printf("%s client <= seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, string(e.Msg.Data))
				return nil
			}),
		}

		dispatcher := EventDispatcher{
			Transceiver:  transceiver,
			EventHandler: generic.CastDelegate1(ctrl.HandleRecvEvent, trans.HandleRecvEvent),
		}

		go func() {
			for {
				for {
					err := ctrl.RequestTime(0)
					if err != nil {
						panic(err)
					}

					fmt.Println(time.Now().Format(time.RFC3339Nano), "client => request time")

					time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
				}
			}
		}()

		go func() {
			for {
				ds := fmt.Sprintf("hello world, %d", rand.Uint64())

				err := trans.SendData([]byte(ds))
				if err != nil {
					panic(err)
				}

				fmt.Println(time.Now().Format(time.RFC3339Nano), "client =>", ds)

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		go func() {
			for {
				for {
					err := ctrl.SendPing()
					if err != nil {
						panic(err)
					}

					fmt.Println(time.Now().Format(time.RFC3339Nano), "client => ping")

					time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
				}
			}
		}()

		dispatcher.Run(context.Background(), generic.CastDelegateVoid1(func(err error) {
			fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= err:", err)
		}))
	}()

	wg.Wait()
}

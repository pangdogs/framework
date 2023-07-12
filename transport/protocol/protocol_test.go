package protocol

import (
	"fmt"
	"golang.org/x/net/context"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
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
					Conn:    conn,
					Encoder: &codec.Encoder{},
					Decoder: &codec.Decoder{
						MsgCreator: codec.DefaultMsgCreator(),
					},
					SendSeq: rand.Uint32(),
					RecvSeq: rand.Uint32(),
				}

				handshake := &HandshakeProtocol{
					Transceiver: transceiver,
				}

				err = handshake.ServerHello(func(e Event[*transport.MsgHello]) (Event[*transport.MsgHello], error) {
					fmt.Println("server: recv hello")
					return Event[*transport.MsgHello]{Flags: transport.Flags(transport.Flag_HelloDone), Msg: &transport.MsgHello{}}, nil
				})
				if err != nil {
					panic(err)
				}

				err = handshake.ServerFinished(Event[*transport.MsgFinished]{
					Msg: &transport.MsgFinished{
						SendSeq: transceiver.SendSeq - 1,
						RecvSeq: transceiver.RecvSeq + 1,
					},
				})
				if err != nil {
					panic(err)
				}

				ctrl := &CtrlProtocol{
					Transceiver:   transceiver,
					RecvRst:       nil,
					RecvSyncTime:  nil,
					RecvHeartbeat: nil,
				}

				trans := &TransProtocol{
					Transceiver: transceiver,
					RecvPayload: func(e Event[*transport.MsgPayload]) error {
						fmt.Println("server: recv", e.Seq, string(e.Msg.Data))
						return nil
					},
				}

				dispather := Dispatcher{
					Transceiver: transceiver,
					Handlers: map[transport.MsgId]Handler{
						transport.MsgId_Rst:       ctrl,
						transport.MsgId_Heartbeat: ctrl,
						transport.MsgId_SyncTime:  ctrl,
						transport.MsgId_Payload:   trans,
					},
				}

				go func() {
					for {
						err := trans.SendPayload(Event[*transport.MsgPayload]{
							Msg: &transport.MsgPayload{
								Data: []byte(fmt.Sprintf("hello world, %d", rand.Uint64())),
							},
						})
						if err != nil {
							panic(err)
						}

						time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
					}
				}()

				dispather.Run(context.Background(), func(err error) bool {
					fmt.Println("server:", err)
					return true
				})
			}()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", "127.0.0.1:7000")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		encoder := &codec.Encoder{
			Encryption:     false,
			PatchMAC:       false,
			CompressedSize: -1,
		}

		decoder := &codec.Decoder{
			MsgCreator: codec.DefaultMsgCreator(),
		}

		transceiver := &Transceiver{
			Conn:    conn,
			Encoder: encoder,
			Decoder: decoder,
		}

		handshake := &HandshakeProtocol{
			Transceiver: transceiver,
		}

		err = handshake.ClientHello(Event[*transport.MsgHello]{Msg: &transport.MsgHello{}}, func(e Event[*transport.MsgHello]) error {
			fmt.Println("client: recv hello")
			return nil
		})
		if err != nil {
			panic(err)
		}

		err = handshake.ClientFinished(func(e Event[*transport.MsgFinished]) error {
			transceiver.RecvSeq = e.Msg.SendSeq
			transceiver.SendSeq = e.Msg.RecvSeq
			fmt.Println("client: recv finished", e.Msg.SendSeq, e.Msg.RecvSeq)
			return nil
		})
		if err != nil {
			panic(err)
		}

		ctrl := &CtrlProtocol{
			Transceiver:   transceiver,
			RecvRst:       nil,
			RecvSyncTime:  nil,
			RecvHeartbeat: nil,
		}

		trans := &TransProtocol{
			Transceiver: transceiver,
			RecvPayload: func(e Event[*transport.MsgPayload]) error {
				fmt.Println("client: recv", e.Seq, string(e.Msg.Data))
				return nil
			},
		}

		dispather := Dispatcher{
			Transceiver: transceiver,
			Handlers: map[transport.MsgId]Handler{
				transport.MsgId_Rst:       ctrl,
				transport.MsgId_Heartbeat: ctrl,
				transport.MsgId_SyncTime:  ctrl,
				transport.MsgId_Payload:   trans,
			},
		}

		go func() {
			for {
				err := trans.SendPayload(Event[*transport.MsgPayload]{
					Msg: &transport.MsgPayload{
						Data: []byte(fmt.Sprintf("hello world, %d", rand.Uint64())),
					},
				})
				if err != nil {
					panic(err)
				}

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		dispather.Run(context.Background(), func(err error) bool {
			fmt.Println("client:", err)
			return true
		})
	}()

	wg.Wait()
}

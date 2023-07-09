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
				}

				handshake := &HandshakeProtocol{
					Transceiver: transceiver,
				}

				err = handshake.ServerHello(func(e Event[*transport.MsgHello]) (Event[*transport.MsgHello], error) {
					fmt.Println("recv client hello")
					return Event[*transport.MsgHello]{Flags: transport.Flags(transport.Flag_HelloDone), Msg: &transport.MsgHello{}}, nil
				})
				if err != nil {
					panic(err)
				}

				sendSeq := rand.Uint32()
				recvSeq := rand.Uint32()

				err = handshake.ServerFinished(Event[*transport.MsgFinished]{
					Msg: &transport.MsgFinished{
						SendSeq: sendSeq,
						RecvSeq: recvSeq,
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
						fmt.Println(e.Msg.Seq, string(e.Msg.Data))
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

				dispather.Run(context.Background(), func(err error) bool {
					fmt.Println(err)
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
			fmt.Println("recv server hello")
			return nil
		})
		if err != nil {
			panic(err)
		}

		var sendSeq, recvSeq uint32

		err = handshake.ClientFinished(func(e Event[*transport.MsgFinished]) error {
			sendSeq = e.Msg.SendSeq
			recvSeq = e.Msg.RecvSeq
			fmt.Println("recv server finished", sendSeq, recvSeq)
			return nil
		})
		if err != nil {
			panic(err)
		}

		trans := &TransProtocol{
			Transceiver: transceiver,
		}

		for {
			sendSeq++

			err := trans.SendPayload(Event[*transport.MsgPayload]{
				Msg: &transport.MsgPayload{
					Seq:  sendSeq,
					Data: []byte("hello world"),
				},
			})
			if err != nil {
				panic(err)
			}

			time.Sleep(5 * time.Second)
		}
	}()

	wg.Wait()
}

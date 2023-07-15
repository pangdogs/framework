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
					Conn:    conn,
					Encoder: &codec.Encoder{},
					Decoder: &codec.Decoder{
						MsgCreator: codec.DefaultMsgCreator(),
					},
					SequencedBuff: SequencedBuff{
						SendSeq: rand.Uint32(),
						RecvSeq: rand.Uint32(),
						Cap:     1024,
					},
				}

				handshake := &HandshakeProtocol{
					Transceiver: transceiver,
				}

				err = handshake.ServerHello(func(e Event[*transport.MsgHello]) (Event[*transport.MsgHello], error) {
					fmt.Println(time.Now().Format(time.RFC3339), "server => recv hello")
					return Event[*transport.MsgHello]{Flags: transport.Flags(transport.Flag_HelloDone), Msg: &transport.MsgHello{}}, nil
				})
				if err != nil {
					panic(err)
				}

				err = handshake.ServerFinished(Event[*transport.MsgFinished]{
					Msg: &transport.MsgFinished{
						Seq: transceiver.SequencedBuff.SendSeq,
						Ack: transceiver.SequencedBuff.RecvSeq,
					},
				})
				if err != nil {
					panic(err)
				}

				ctrl := &CtrlProtocol{
					Transceiver: transceiver,
				}

				trans := &TransProtocol{
					Transceiver: transceiver,
					RecvPayload: func(e Event[*transport.MsgPayload]) error {
						fmt.Printf("%s server => recv seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339), e.Seq, e.Ack, string(e.Msg.Data))
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
						ds := fmt.Sprintf("hello world, %d", rand.Uint64())

						err := trans.SendPayload(Event[*transport.MsgPayload]{
							Flags: transport.Flags(transport.Flag_Sequenced),
							Msg: &transport.MsgPayload{
								Data: []byte(ds),
							},
						})
						if err != nil {
							panic(err)
						}

						fmt.Println(time.Now().Format(time.RFC3339), "server => send", ds)

						time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
					}
				}()

				dispather.Run(context.Background(), func(err error) bool {
					fmt.Println(time.Now().Format(time.RFC3339), "server => err", err)
					return true
				})
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
			Conn:    conn,
			Encoder: &codec.Encoder{},
			Decoder: &codec.Decoder{
				MsgCreator: codec.DefaultMsgCreator(),
			},
			SequencedBuff: SequencedBuff{
				Cap: 1024,
			},
		}

		handshake := &HandshakeProtocol{
			Transceiver: transceiver,
		}

		err = handshake.ClientHello(Event[*transport.MsgHello]{Msg: &transport.MsgHello{}}, func(e Event[*transport.MsgHello]) error {
			fmt.Println(time.Now().Format(time.RFC3339), "client => recv hello")
			return nil
		})
		if err != nil {
			panic(err)
		}

		err = handshake.ClientFinished(func(e Event[*transport.MsgFinished]) error {
			transceiver.SequencedBuff.SendSeq = e.Msg.Ack
			transceiver.SequencedBuff.RecvSeq = e.Msg.Seq
			fmt.Println(time.Now().Format(time.RFC3339), "client => recv finished", e.Msg.Seq, e.Msg.Ack)
			return nil
		})
		if err != nil {
			panic(err)
		}

		ctrl := &CtrlProtocol{
			Transceiver: transceiver,
		}

		trans := &TransProtocol{
			Transceiver: transceiver,
			RecvPayload: func(e Event[*transport.MsgPayload]) error {
				fmt.Printf("%s client => recv seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339), e.Seq, e.Ack, string(e.Msg.Data))
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
				ds := fmt.Sprintf("hello world, %d", rand.Uint64())

				err := trans.SendPayload(Event[*transport.MsgPayload]{
					Flags: transport.Flags(transport.Flag_Sequenced),
					Msg: &transport.MsgPayload{
						Data: []byte(fmt.Sprintf("hello world, %d", rand.Uint64())),
					},
				})
				if err != nil {
					panic(err)
				}

				fmt.Println(time.Now().Format(time.RFC3339), "client => send", ds)

				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
			}
		}()

		dispather.Run(context.Background(), func(err error) bool {
			fmt.Println(time.Now().Format(time.RFC3339), "client => err ", err)
			return true
		})
	}()

	wg.Wait()
}

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
						Cap: 1024,
					},
				}
				transceiver.SequencedBuff.Reset(rand.Uint32(), rand.Uint32())

				handshake := &HandshakeProtocol{
					Transceiver: transceiver,
				}

				err = handshake.ServerHello(func(e Event[*transport.MsgHello]) (Event[*transport.MsgHello], error) {
					fmt.Println(time.Now().Format(time.RFC3339Nano), "server <= hello")
					return Event[*transport.MsgHello]{Flags: transport.Flags(transport.Flag_HelloDone), Msg: &transport.MsgHello{}}, nil
				})
				if err != nil {
					panic(err)
				}

				err = handshake.ServerFinished(Event[*transport.MsgFinished]{
					Msg: &transport.MsgFinished{
						SendSeq: transceiver.SequencedBuff.SendSeq,
						RecvSeq: transceiver.SequencedBuff.RecvSeq,
					},
				})
				if err != nil {
					panic(err)
				}

				ctrl := &CtrlProtocol{
					Transceiver: transceiver,
					HandleHeartbeat: func(e Event[*transport.MsgHeartbeat]) error {
						text := "ping"
						if e.Flags.Is(transport.Flag_Pong) {
							text = "pong"
						}
						fmt.Printf("%s server <= seq:%d ack:%d %s\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, text)
						return nil
					},
				}

				trans := &TransProtocol{
					Transceiver: transceiver,
					HandlePayload: func(e Event[*transport.MsgPayload]) error {
						fmt.Printf("%s server <= seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, string(e.Msg.Data))
						return nil
					},
				}

				dispatcher := EventDispatcher{
					Transceiver: transceiver,
					ErrorHandler: func(ctx context.Context, err error) {
						fmt.Println(time.Now().Format(time.RFC3339Nano), "server <= err", err)
					},
				}
				dispatcher.Add(ctrl)
				dispatcher.Add(trans)

				go func() {
					for {
						ds := fmt.Sprintf("hello world, %d", rand.Uint64())

						err := trans.SendData([]byte(ds), true)
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

				go func() {
					for {
						for {
							err := ctrl.SendSyncTime()
							if err != nil {
								panic(err)
							}

							fmt.Println(time.Now().Format(time.RFC3339Nano), "server => sync time")

							time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
						}
					}
				}()

				dispatcher.Run(context.Background())
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
			fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= hello")
			return nil
		})
		if err != nil {
			panic(err)
		}

		var sendSeq, recvSeq uint32

		err = handshake.ClientFinished(func(e Event[*transport.MsgFinished]) error {
			sendSeq = e.Msg.RecvSeq
			recvSeq = e.Msg.SendSeq
			fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= finished", e.Msg.SendSeq, e.Msg.RecvSeq)
			return nil
		})
		if err != nil {
			panic(err)
		}

		transceiver.SequencedBuff.Reset(sendSeq, recvSeq)

		ctrl := &CtrlProtocol{
			Transceiver: transceiver,
			HandleHeartbeat: func(e Event[*transport.MsgHeartbeat]) error {
				text := "ping"
				if e.Flags.Is(transport.Flag_Pong) {
					text = "pong"
				}
				fmt.Printf("%s client <= seq:%d ack:%d %s\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, text)
				return nil
			},
			HandleSyncTime: func(e Event[*transport.MsgSyncTime]) error {
				fmt.Printf("%s client <= sync time %d\n", time.Now().Format(time.RFC3339Nano), e.Msg.UnixMilli)
				return nil
			},
		}

		trans := &TransProtocol{
			Transceiver: transceiver,
			HandlePayload: func(e Event[*transport.MsgPayload]) error {
				fmt.Printf("%s client <= seq:%d ack:%d data:%q\n", time.Now().Format(time.RFC3339Nano), e.Seq, e.Ack, string(e.Msg.Data))
				return nil
			},
		}

		dispatcher := EventDispatcher{
			Transceiver: transceiver,
			ErrorHandler: func(ctx context.Context, err error) {
				fmt.Println(time.Now().Format(time.RFC3339Nano), "client <= err", err)
			},
		}
		dispatcher.Add(ctrl)
		dispatcher.Add(trans)

		go func() {
			for {
				ds := fmt.Sprintf("hello world, %d", rand.Uint64())

				err := trans.SendData([]byte(ds), true)
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

		dispatcher.Run(context.Background())
	}()

	wg.Wait()
}

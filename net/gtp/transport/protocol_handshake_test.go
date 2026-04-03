package transport

import (
	"context"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestHandshakeHelloAndFinished(t *testing.T) {
	clientTr, serverTr := newUnsequencedPipeTransceivers(t)
	client := &HandshakeProtocol{Transceiver: clientTr}
	server := &HandshakeProtocol{Transceiver: serverTr}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ServerHello(context.Background(), func(e Event[*gtp.MsgHello]) (Event[*gtp.MsgHello], error) {
			return Event[*gtp.MsgHello]{
				Flags: gtp.Flags(gtp.Flag_HelloDone),
				Msg:   &gtp.MsgHello{SessionId: "srv"},
			}, nil
		})
	}()

	err := client.ClientHello(context.Background(), Event[*gtp.MsgHello]{Msg: &gtp.MsgHello{SessionId: "cli"}}, func(e Event[*gtp.MsgHello]) error {
		if e.Msg.SessionId != "srv" {
			t.Fatalf("unexpected hello reply: %+v", e)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ClientHello failed: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("ServerHello failed: %v", err)
	}

	serverErr = make(chan error, 1)
	go func() {
		serverErr <- server.ServerFinished(context.Background(), Event[*gtp.MsgFinished]{Msg: &gtp.MsgFinished{SendSeq: 1, RecvSeq: 2}})
	}()

	if err := client.ClientFinished(context.Background(), func(e Event[*gtp.MsgFinished]) error {
		if e.Msg.SendSeq != 1 || e.Msg.RecvSeq != 2 {
			t.Fatalf("unexpected finished event: %+v", e)
		}
		return nil
	}); err != nil {
		t.Fatalf("ClientFinished failed: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("ServerFinished failed: %v", err)
	}
}

func TestHandshakeNilArgsAndUnexpectedMessages(t *testing.T) {
	clientTr, serverTr := newUnsequencedPipeTransceivers(t)
	client := &HandshakeProtocol{Transceiver: clientTr}
	server := &HandshakeProtocol{Transceiver: serverTr}

	if err := client.ClientHello(context.Background(), Event[*gtp.MsgHello]{Msg: &gtp.MsgHello{}}, nil); err == nil {
		t.Fatal("expected helloFin error")
	}
	if err := server.ServerHello(context.Background(), nil); err == nil {
		t.Fatal("expected helloAccept error")
	}
	if err := client.ClientFinished(context.Background(), nil); err == nil {
		t.Fatal("expected finishedAccept error")
	}

	go func() {
		_ = clientTr.Send(newPayloadEvent("unexpected"))
	}()
	if err := server.ServerHello(context.Background(), func(e Event[*gtp.MsgHello]) (Event[*gtp.MsgHello], error) {
		return Event[*gtp.MsgHello]{Msg: &gtp.MsgHello{}}, nil
	}); err == nil {
		t.Fatal("expected unexpected message error")
	}
}

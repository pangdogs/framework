package transport

import (
	"testing"

	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/framework/net/gtp"
)

func TestCtrlProtocolSendMethods(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)
	ctrl := &CtrlProtocol{Transceiver: client}

	done := make(chan error, 1)
	go func() { done <- ctrl.RequestTime(7) }()
	e := recvWithTimeout(t, server)
	if err := <-done; err != nil {
		t.Fatalf("RequestTime failed: %v", err)
	}
	syncTime := AssertEvent[*gtp.MsgSyncTime](e)
	if !syncTime.Flags.Is(gtp.Flag_ReqTime) || syncTime.Msg.CorrId != 7 {
		t.Fatalf("unexpected sync time event: %+v", syncTime)
	}

	go func() { done <- ctrl.SendPing() }()
	e = recvWithTimeout(t, server)
	if err := <-done; err != nil {
		t.Fatalf("SendPing failed: %v", err)
	}
	hb := AssertEvent[*gtp.MsgHeartbeat](e)
	if !hb.Flags.Is(gtp.Flag_Ping) {
		t.Fatalf("unexpected heartbeat flags: %v", hb.Flags)
	}
}

func TestCtrlProtocolHandleEvent(t *testing.T) {
	client, server := newUnsequencedPipeTransceivers(t)
	ctrl := &CtrlProtocol{Transceiver: client}

	rstCalled := false
	syncCalled := false
	hbCalled := false
	ctrl.RstHandler = generic.CastDelegateVoid1(func(Event[*gtp.MsgRst]) { rstCalled = true })
	ctrl.SyncTimeHandler = generic.CastDelegateVoid1(func(Event[*gtp.MsgSyncTime]) { syncCalled = true })
	ctrl.HeartbeatHandler = generic.CastDelegateVoid1(func(Event[*gtp.MsgHeartbeat]) { hbCalled = true })

	ctrl.HandleEvent(Event[*gtp.MsgRst]{Msg: &gtp.MsgRst{Message: "x"}}.Interface())
	if !rstCalled {
		t.Fatal("expected rst handler")
	}

	respCh := make(chan IEvent, 1)
	go func() {
		respCh <- recvWithTimeout(t, server)
	}()
	ctrl.HandleEvent(Event[*gtp.MsgSyncTime]{
		Flags: gtp.Flags(gtp.Flag_ReqTime),
		Msg:   &gtp.MsgSyncTime{CorrId: 9, LocalTime: 1},
	}.Interface())
	if !syncCalled {
		t.Fatal("expected sync time handler")
	}
	resp := <-respCh
	if !AssertEvent[*gtp.MsgSyncTime](resp).Flags.Is(gtp.Flag_RespTime) {
		t.Fatal("expected response time event")
	}

	pongCh := make(chan IEvent, 1)
	go func() {
		pongCh <- recvWithTimeout(t, server)
	}()
	ctrl.HandleEvent(Event[*gtp.MsgHeartbeat]{
		Flags: gtp.Flags(gtp.Flag_Ping),
		Msg:   &gtp.MsgHeartbeat{},
	}.Interface())
	if !hbCalled {
		t.Fatal("expected heartbeat handler")
	}
	pong := <-pongCh
	if !AssertEvent[*gtp.MsgHeartbeat](pong).Flags.Is(gtp.Flag_Pong) {
		t.Fatal("expected pong event")
	}
}

func TestCtrlProtocolNilTransceiverErrors(t *testing.T) {
	ctrl := &CtrlProtocol{}
	if err := ctrl.SendRst(nil); err == nil {
		t.Fatal("expected SendRst error")
	}
	if err := ctrl.RequestTime(1); err == nil {
		t.Fatal("expected RequestTime error")
	}
	if err := ctrl.SendPing(); err == nil {
		t.Fatal("expected SendPing error")
	}
}

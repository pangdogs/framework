package gtp

import (
	"errors"
	"testing"
)

func TestNewMsgCreatorDeclareAndNew(t *testing.T) {
	creator := NewMsgCreator()

	if _, err := creator.New(MsgID_Hello); !errors.Is(err, ErrNotDeclared) {
		t.Fatalf("expected ErrNotDeclared, got %v", err)
	}

	creator.Declare(&MsgPayload{})

	msg, err := creator.New(MsgID_Payload)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if _, ok := msg.(*MsgPayload); !ok {
		t.Fatalf("unexpected message type: %T", msg)
	}
	if msg == (&MsgPayload{}) {
		t.Fatal("expected newly allocated message")
	}
}

func TestMsgCreatorDeclarePanics(t *testing.T) {
	creator := NewMsgCreator()

	assertPanic(t, func() {
		creator.Declare(nil)
	})

	creator.Declare(&MsgPayload{})
	assertPanic(t, func() {
		creator.Declare(&MsgPayload{})
	})
}

func TestDefaultMsgCreatorBuiltins(t *testing.T) {
	builtin := []MsgID{
		MsgID_Hello,
		MsgID_ECDHESecretKeyExchange,
		MsgID_ChangeCipherSpec,
		MsgID_Auth,
		MsgID_Continue,
		MsgID_Finished,
		MsgID_Rst,
		MsgID_Heartbeat,
		MsgID_SyncTime,
		MsgID_Payload,
	}

	for _, msgID := range builtin {
		msg, err := DefaultMsgCreator().New(msgID)
		if err != nil {
			t.Fatalf("DefaultMsgCreator.New(%d) failed: %v", msgID, err)
		}
		if msg == nil || msg.MsgID() != msgID {
			t.Fatalf("unexpected builtin message for %d: %#v", msgID, msg)
		}
	}
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

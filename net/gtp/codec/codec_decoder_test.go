package codec

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewDecoderPanicsWithNilCreator(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	NewDecoder(nil)
}

func TestDecoderDecodePlainPacket(t *testing.T) {
	encoder := NewEncoder()
	buf := mustEncode(t, encoder, newTestPayload())
	defer buf.Release()

	decoder := NewDecoder(gtp.DefaultMsgCreator())
	mp, n, err := decoder.Decode(buf.Payload(), nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if n != len(buf.Payload()) {
		t.Fatalf("unexpected consumed length: got %d want %d", n, len(buf.Payload()))
	}
	if !bytes.Equal(mp.Msg.(*gtp.MsgPayload).Data, newTestPayload().Data) {
		t.Fatal("unexpected decoded payload")
	}
}

func TestDecoderDecodeErrors(t *testing.T) {
	t.Run("nil msg creator", func(t *testing.T) {
		var decoder Decoder
		if _, _, err := decoder.Decode([]byte{1, 2, 3}, nil); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("peek length short buffer", func(t *testing.T) {
		decoder := NewDecoder(gtp.DefaultMsgCreator())
		if _, _, err := decoder.Decode([]byte{1, 2, 3}, nil); !errors.Is(err, ErrUnableToPeekLength) {
			t.Fatalf("expected ErrUnableToPeekLength, got %v", err)
		}
	})

	t.Run("packet short buffer", func(t *testing.T) {
		buf := mustEncode(t, NewEncoder(), newTestPayload())
		defer buf.Release()

		decoder := NewDecoder(gtp.DefaultMsgCreator())
		_, length, err := decoder.Decode(buf.Payload()[:len(buf.Payload())-1], nil)
		if !errors.Is(err, io.ErrShortBuffer) {
			t.Fatalf("expected short buffer error, got %v", err)
		}
		if length != len(buf.Payload()) {
			t.Fatalf("unexpected peeked length: got %d want %d", length, len(buf.Payload()))
		}
	})

	t.Run("validation", func(t *testing.T) {
		buf := mustEncode(t, NewEncoder(), newTestPayload())
		defer buf.Release()

		decoder := NewDecoder(gtp.DefaultMsgCreator())
		if _, _, err := decoder.Decode(buf.Payload(), stubValidation{err: errTest}); !errors.Is(err, errTest) {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("new msg error", func(t *testing.T) {
		buf := mustEncode(t, NewEncoder(), newTestPayload())
		defer buf.Release()

		decoder := NewDecoder(stubMsgCreator{
			newFn: func(gtp.MsgId) (gtp.Msg, error) { return nil, errTest },
		})
		if _, _, err := decoder.Decode(buf.Payload(), nil); !errors.Is(err, errTest) {
			t.Fatalf("expected new msg error, got %v", err)
		}
	})

	t.Run("msg write error", func(t *testing.T) {
		buf := mustEncode(t, NewEncoder(), newTestPayload())
		defer buf.Release()

		decoder := NewDecoder(stubMsgCreator{
			newFn: func(gtp.MsgId) (gtp.Msg, error) { return &failingMsg{writeErr: errTest}, nil },
		})
		if _, _, err := decoder.Decode(buf.Payload(), nil); !errors.Is(err, errTest) {
			t.Fatalf("expected msg write error, got %v", err)
		}
	})
}

func TestDecoderMissingModules(t *testing.T) {
	encEncryption, decEncryption := newTestEncryptionPair(t)
	encoder := NewEncoder().
		SetCompression(newTestCompression(t), 1).
		SetAuthentication(NewAuthentication(newTestHMAC(t))).
		SetEncryption(encEncryption)

	buf := mustEncode(t, encoder, newTestPayload())
	defer buf.Release()

	t.Run("missing encryption", func(t *testing.T) {
		decoder := NewDecoder(gtp.DefaultMsgCreator())
		if _, _, err := decoder.Decode(buf.Payload(), nil); err == nil {
			t.Fatal("expected encryption error")
		}
	})

	t.Run("missing authentication", func(t *testing.T) {
		decoder := NewDecoder(gtp.DefaultMsgCreator()).
			SetEncryption(decEncryption).
			SetCompression(newTestCompression(t), 1<<20)
		if _, _, err := decoder.Decode(buf.Payload(), nil); err == nil {
			t.Fatal("expected authentication error")
		}
	})

	t.Run("missing compression", func(t *testing.T) {
		decoder := NewDecoder(gtp.DefaultMsgCreator()).
			SetEncryption(decEncryption).
			SetAuthentication(NewAuthentication(newTestHMAC(t)))
		if _, _, err := decoder.Decode(buf.Payload(), nil); err == nil {
			t.Fatal("expected compression error")
		}
	})
}

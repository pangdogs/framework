package codec

import (
	"bytes"
	"errors"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

func TestNewEncoder(t *testing.T) {
	if NewEncoder() == nil {
		t.Fatal("expected encoder")
	}
}

func TestEncoderEncodeNilMsg(t *testing.T) {
	_, err := NewEncoder().Encode(gtp.Flags_None(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncoderEncodePlainPacket(t *testing.T) {
	encoder := NewEncoder()
	msg := newTestPayload()

	buf := mustEncode(t, encoder, msg)
	defer buf.Release()

	var head gtp.MsgHead
	if _, err := head.Write(buf.Payload()); err != nil {
		t.Fatalf("head.Write failed: %v", err)
	}
	if head.Len != uint32(len(buf.Payload())) {
		t.Fatalf("unexpected packet length: got %d want %d", head.Len, len(buf.Payload()))
	}
	if head.Flags != gtp.Flags_None() {
		t.Fatalf("unexpected flags: %v", head.Flags)
	}

	decoder := NewDecoder(gtp.DefaultMsgCreator())
	mp, _, err := decoder.Decode(buf.Payload(), nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	got := mp.Body.(*gtp.MsgPayload)
	if !bytes.Equal(got.Data, msg.Data) {
		t.Fatal("unexpected decoded payload")
	}
}

func TestEncoderEncodeWithModules(t *testing.T) {
	encEncryption, decEncryption := newTestEncryptionPair(t)
	encoder := NewEncoder().
		SetCompression(newTestCompression(t), 1).
		SetAuthentication(NewAuthentication(newTestHMAC(t))).
		SetEncryption(encEncryption)

	buf := mustEncode(t, encoder, newTestPayload())
	defer buf.Release()

	var head gtp.MsgHead
	if _, err := head.Write(buf.Payload()); err != nil {
		t.Fatalf("head.Write failed: %v", err)
	}
	if !head.Flags.Is(gtp.Flag_Encrypted) || !head.Flags.Is(gtp.Flag_Signed) || !head.Flags.Is(gtp.Flag_Compressed) {
		t.Fatalf("unexpected flags: %v", head.Flags)
	}

	decoder := NewDecoder(gtp.DefaultMsgCreator()).
		SetCompression(newTestCompression(t), 1<<20).
		SetAuthentication(NewAuthentication(newTestHMAC(t))).
		SetEncryption(decEncryption)

	mp, _, err := decoder.Decode(buf.Payload(), nil)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if got := mp.Body.(*gtp.MsgPayload); !bytes.Equal(got.Data, newTestPayload().Data) {
		t.Fatal("unexpected decoded payload")
	}
}

func TestEncoderModuleErrors(t *testing.T) {
	msg := newTestPayload()

	t.Run("encryption size", func(t *testing.T) {
		encoder := NewEncoder().SetEncryption(stubEncryption{
			sizeOfAdditionFn: func(int) (int, error) { return 0, errTest },
		})
		if _, err := encoder.Encode(gtp.Flags_None(), msg); !errors.Is(err, errTest) {
			t.Fatalf("expected encryption size error, got %v", err)
		}
	})

	t.Run("authentication size", func(t *testing.T) {
		encoder := NewEncoder().
			SetEncryption(stubEncryption{}).
			SetAuthentication(stubAuthentication{
				sizeOfAdditionFn: func(int) (int, error) { return 0, errTest },
			})
		if _, err := encoder.Encode(gtp.Flags_None(), msg); !errors.Is(err, errTest) {
			t.Fatalf("expected authentication size error, got %v", err)
		}
	})

	t.Run("compression", func(t *testing.T) {
		encoder := NewEncoder().SetCompression(stubCompression{
			compressFn: func([]byte) (binaryutil.Bytes, bool, error) { return binaryutil.EmptyBytes, false, errTest },
		}, 1)
		if _, err := encoder.Encode(gtp.Flags_None(), msg); !errors.Is(err, errTest) {
			t.Fatalf("expected compression error, got %v", err)
		}
	})

	t.Run("authentication sign", func(t *testing.T) {
		encoder := NewEncoder().
			SetEncryption(stubEncryption{}).
			SetAuthentication(stubAuthentication{
				signFn: func(gtp.MsgId, gtp.Flags, []byte) (binaryutil.Bytes, error) { return binaryutil.EmptyBytes, errTest },
			})
		if _, err := encoder.Encode(gtp.Flags_None(), msg); !errors.Is(err, errTest) {
			t.Fatalf("expected sign error, got %v", err)
		}
	})

	t.Run("encryption transform", func(t *testing.T) {
		encoder := NewEncoder().SetEncryption(stubEncryption{
			transformFn: func([]byte, []byte) (binaryutil.Bytes, error) { return binaryutil.EmptyBytes, errTest },
		})
		if _, err := encoder.Encode(gtp.Flags_None(), msg); !errors.Is(err, errTest) {
			t.Fatalf("expected transform error, got %v", err)
		}
	})
}

package codec

import (
	"bytes"
	"errors"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

func TestNewAuthenticationPanicsWithNilHMAC(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	NewAuthentication(nil)
}

func TestAuthenticationSignAndAuthRoundTrip(t *testing.T) {
	auth := &Authentication{HMAC: newTestHMAC(t)}
	msg := []byte("hello-auth")

	signed, err := auth.Sign(gtp.MsgId_Payload, gtp.Flags_None().Setd(gtp.Flag_Encrypted, true), msg)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	defer signed.Release()

	got, err := auth.Auth(gtp.MsgId_Payload, gtp.Flags_None().Setd(gtp.Flag_Encrypted, true), signed.Payload())
	if err != nil {
		t.Fatalf("Auth failed: %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatalf("unexpected auth payload: got %q want %q", got, msg)
	}
}

func TestAuthenticationAuthInvalidMAC(t *testing.T) {
	auth := &Authentication{HMAC: newTestHMAC(t)}

	signed, err := auth.Sign(gtp.MsgId_Payload, gtp.Flags_None(), []byte("hello"))
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	defer signed.Release()

	payload := bytes.Clone(signed.Payload())
	payload[len(payload)-1] ^= 0xFF

	_, err = auth.Auth(gtp.MsgId_Payload, gtp.Flags_None(), payload)
	if !errors.Is(err, ErrInvalidMAC) {
		t.Fatalf("expected ErrInvalidMAC, got %v", err)
	}
}

func TestAuthenticationNilHMACErrors(t *testing.T) {
	auth := &Authentication{}

	if _, err := auth.Sign(gtp.MsgId_Payload, gtp.Flags_None(), []byte("x")); err == nil {
		t.Fatal("expected Sign error")
	}
	if _, err := auth.Auth(gtp.MsgId_Payload, gtp.Flags_None(), []byte("x")); err == nil {
		t.Fatal("expected Auth error")
	}
	if _, err := auth.SizeOfAddition(1); err == nil {
		t.Fatal("expected SizeOfAddition error")
	}
}

func TestAuthenticationSizeOfAddition(t *testing.T) {
	auth := &Authentication{HMAC: newTestHMAC(t)}

	size, err := auth.SizeOfAddition(123)
	if err != nil {
		t.Fatalf("SizeOfAddition failed: %v", err)
	}

	want := binaryutil.SizeofVarint(123) + binaryutil.SizeofVarint(int64(auth.HMAC.Size())) + auth.HMAC.Size()
	if size != want {
		t.Fatalf("unexpected addition size: got %d want %d", size, want)
	}
}

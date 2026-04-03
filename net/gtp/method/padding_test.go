package method

import (
	"bytes"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewPadding(t *testing.T) {
	cases := []gtp.PaddingMode{
		gtp.PaddingMode_Pkcs7,
		gtp.PaddingMode_X923,
	}

	for _, tc := range cases {
		t.Run(tc.String(), func(t *testing.T) {
			padding, err := NewPadding(tc)
			if err != nil {
				t.Fatalf("NewPadding failed: %v", err)
			}

			buf := make([]byte, 8)
			copy(buf, []byte("abc"))
			if err := padding.Pad(buf, 3); err != nil {
				t.Fatalf("Pad failed: %v", err)
			}
			got, err := padding.Unpad(buf)
			if err != nil {
				t.Fatalf("Unpad failed: %v", err)
			}
			if !bytes.Equal(got, []byte("abc")) {
				t.Fatalf("unexpected unpadded payload: %q", got)
			}
		})
	}
}

func TestNewPaddingInvalid(t *testing.T) {
	if _, err := NewPadding(gtp.PaddingMode(255)); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
}

func TestPaddingErrors(t *testing.T) {
	t.Run("pkcs7 wrong pad len", func(t *testing.T) {
		if err := (_Pkcs7{}).Pad(make([]byte, 3), 3); err == nil {
			t.Fatal("expected Pad error")
		}
	})

	t.Run("x923 wrong pad len", func(t *testing.T) {
		if err := (_X923{}).Pad(make([]byte, 3), 3); err == nil {
			t.Fatal("expected Pad error")
		}
	})

	t.Run("pkcs7 invalid padding", func(t *testing.T) {
		if _, err := (_Pkcs7{}).Unpad([]byte{1, 2, 3, 2}); err == nil {
			t.Fatal("expected Unpad error")
		}
	})

	t.Run("x923 invalid padding", func(t *testing.T) {
		if _, err := (_X923{}).Unpad([]byte{1, 2, 3, 2}); err == nil {
			t.Fatal("expected Unpad error")
		}
	})
}

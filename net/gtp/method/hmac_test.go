package method

import (
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewHMAC(t *testing.T) {
	key := []byte("secret-key")
	cases := []struct {
		name string
		hash gtp.Hash
	}{
		{name: "sha256", hash: gtp.Hash_SHA256},
		{name: "blake2b256", hash: gtp.Hash_BLAKE2b256},
		{name: "blake2b384", hash: gtp.Hash_BLAKE2b384},
		{name: "blake2b512", hash: gtp.Hash_BLAKE2b512},
		{name: "blake2s256", hash: gtp.Hash_BLAKE2s256},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHMAC(tc.hash, key)
			if err != nil {
				t.Fatalf("NewHMAC failed: %v", err)
			}
			if h.Size() <= 0 {
				t.Fatalf("unexpected hash size: %d", h.Size())
			}
		})
	}
}

func TestNewHMACInvalid(t *testing.T) {
	if _, err := NewHMAC(gtp.Hash(255), []byte("k")); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
}

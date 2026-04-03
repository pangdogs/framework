package codec

import (
	"bytes"
	"errors"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
)

func TestNewEncryptionPanicsOnInvalidArgs(t *testing.T) {
	t.Run("nil cipher", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		NewEncryption(nil, nil, nil)
	})

	t.Run("missing padding", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		NewEncryption(stubCipher{pad: true}, nil, nil)
	})

	t.Run("missing nonce provider", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		NewEncryption(stubCipher{nonceSize: 12}, nil, nil)
	})
}

func TestEncryptionTransformingRoundTripAEAD(t *testing.T) {
	encryptor, decryptor := newTestEncryptionPair(t)
	src := bytes.Repeat([]byte("encrypt-me-"), 16)

	encrypted, err := encryptor.Transforming(nil, src)
	if err != nil {
		t.Fatalf("encrypt Transforming failed: %v", err)
	}
	defer encrypted.Release()

	dst := make([]byte, len(encrypted.Payload()))
	decrypted, err := decryptor.Transforming(dst, encrypted.Payload())
	if err != nil {
		t.Fatalf("decrypt Transforming failed: %v", err)
	}
	defer decrypted.Release()

	if !bytes.Equal(decrypted.Payload(), src) {
		t.Fatal("unexpected decrypted payload")
	}
}

func TestEncryptionTransformingRoundTripWithPadding(t *testing.T) {
	key := bytes.Repeat([]byte{5}, 16)
	iv := bytes.Repeat([]byte{6}, 16)

	encryptCipher, decryptCipher, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_CBC, key, iv, nil)
	if err != nil {
		t.Fatalf("NewCipher failed: %v", err)
	}
	padding, err := method.NewPadding(gtp.PaddingMode_Pkcs7)
	if err != nil {
		t.Fatalf("NewPadding failed: %v", err)
	}

	encryptor := NewEncryption(encryptCipher, padding, nil)
	decryptor := NewEncryption(decryptCipher, padding, nil)

	src := []byte("pad-me")
	encrypted, err := encryptor.Transforming(nil, src)
	if err != nil {
		t.Fatalf("encrypt Transforming failed: %v", err)
	}
	defer encrypted.Release()

	dst := make([]byte, len(encrypted.Payload()))
	decrypted, err := decryptor.Transforming(dst, encrypted.Payload())
	if err != nil {
		t.Fatalf("decrypt Transforming failed: %v", err)
	}
	defer decrypted.Release()

	if !bytes.Equal(decrypted.Payload(), src) {
		t.Fatal("unexpected decrypted payload")
	}
}

func TestEncryptionErrors(t *testing.T) {
	t.Run("nil cipher receiver", func(t *testing.T) {
		e := &Encryption{}
		if _, err := e.Transforming(nil, []byte("x")); err == nil {
			t.Fatal("expected Transforming error")
		}
		if _, err := e.SizeOfAddition(1); err == nil {
			t.Fatal("expected SizeOfAddition error")
		}
	})

	t.Run("fetch nonce error", func(t *testing.T) {
		key := bytes.Repeat([]byte{7}, 16)
		encCipher, _, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, nil, nil)
		if err != nil {
			t.Fatalf("NewCipher failed: %v", err)
		}
		e := &Encryption{
			Cipher:     encCipher,
			FetchNonce: func() ([]byte, error) { return nil, errTest },
		}
		if _, err := e.Transforming(nil, []byte("abc")); !errors.Is(err, errTest) {
			t.Fatalf("expected nonce fetch error, got %v", err)
		}
	})

	t.Run("padding error", func(t *testing.T) {
		e := &Encryption{
			Cipher:  stubCipher{pad: true, inputSize: func(size int) int { return size + 1 }},
			Padding: stubPadding{padFn: func([]byte, int) error { return errTest }},
		}
		if _, err := e.Transforming(nil, []byte("abc")); !errors.Is(err, errTest) {
			t.Fatalf("expected padding error, got %v", err)
		}
	})

	t.Run("cipher error", func(t *testing.T) {
		e := &Encryption{
			Cipher: stubCipher{
				transform: func([]byte, []byte, []byte) (int, error) { return 0, errTest },
			},
		}
		if _, err := e.Transforming(make([]byte, 3), []byte("abc")); !errors.Is(err, errTest) {
			t.Fatalf("expected cipher error, got %v", err)
		}
	})

	t.Run("unpad error", func(t *testing.T) {
		e := &Encryption{
			Cipher:  stubCipher{unpad: true},
			Padding: stubPadding{unpadFn: func([]byte) ([]byte, error) { return nil, errTest }},
		}
		if _, err := e.Transforming(make([]byte, 3), []byte("abc")); !errors.Is(err, errTest) {
			t.Fatalf("expected unpad error, got %v", err)
		}
	})
}

func TestEncryptionSizeOfAddition(t *testing.T) {
	key := bytes.Repeat([]byte{8}, 16)
	encCipher, _, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, nil, nil)
	if err != nil {
		t.Fatalf("NewCipher failed: %v", err)
	}

	e := &Encryption{Cipher: encCipher}
	size, err := e.SizeOfAddition(8)
	if err != nil {
		t.Fatalf("SizeOfAddition failed: %v", err)
	}
	if size <= 0 {
		t.Fatalf("expected positive additional size, got %d", size)
	}
}

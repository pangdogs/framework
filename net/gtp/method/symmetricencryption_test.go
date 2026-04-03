package method

import (
	"bytes"
	"crypto/cipher"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewCipherBlock(t *testing.T) {
	block, err := NewCipherBlock(gtp.SymmetricEncryption_AES, bytes.Repeat([]byte{1}, 16))
	if err != nil {
		t.Fatalf("NewCipherBlock failed: %v", err)
	}
	if block.BlockSize() != 16 {
		t.Fatalf("unexpected block size: %d", block.BlockSize())
	}
}

func TestNewCipherBlockInvalid(t *testing.T) {
	if _, err := NewCipherBlock(gtp.SymmetricEncryption(255), []byte("1234567890123456")); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
	if _, err := NewCipherBlock(gtp.SymmetricEncryption_AES, []byte("short")); err == nil {
		t.Fatal("expected invalid AES key error")
	}
}

func TestNewBlockCipherMode(t *testing.T) {
	block, err := NewCipherBlock(gtp.SymmetricEncryption_AES, bytes.Repeat([]byte{1}, 16))
	if err != nil {
		t.Fatalf("NewCipherBlock failed: %v", err)
	}
	iv := bytes.Repeat([]byte{2}, block.BlockSize())

	cases := []gtp.BlockCipherMode{
		gtp.BlockCipherMode_CTR,
		gtp.BlockCipherMode_CBC,
		gtp.BlockCipherMode_CFB,
		gtp.BlockCipherMode_OFB,
		gtp.BlockCipherMode_GCM,
	}

	for _, tc := range cases {
		t.Run(tc.String(), func(t *testing.T) {
			encryptor, decrypter, err := NewBlockCipherMode(tc, block, iv)
			if err != nil {
				t.Fatalf("NewBlockCipherMode failed: %v", err)
			}
			if encryptor == nil || decrypter == nil {
				t.Fatal("expected encryptor and decrypter")
			}
		})
	}
}

func TestNewBlockCipherModeInvalid(t *testing.T) {
	block, err := NewCipherBlock(gtp.SymmetricEncryption_AES, bytes.Repeat([]byte{1}, 16))
	if err != nil {
		t.Fatalf("NewCipherBlock failed: %v", err)
	}

	if _, _, err := NewBlockCipherMode(gtp.BlockCipherMode(255), block, bytes.Repeat([]byte{2}, block.BlockSize())); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
	if _, _, err := NewBlockCipherMode(gtp.BlockCipherMode_CTR, block, []byte("bad-iv")); err == nil {
		t.Fatal("expected invalid iv error")
	}
}

func TestNewCipherRoundTrip(t *testing.T) {
	t.Run("aes-ctr", func(t *testing.T) {
		key := bytes.Repeat([]byte{1}, 16)
		iv := bytes.Repeat([]byte{2}, 16)
		encryptor, decrypter, err := NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_CTR, key, iv, nil)
		if err != nil {
			t.Fatalf("NewCipher failed: %v", err)
		}

		src := []byte("stream cipher payload")
		dst := make([]byte, len(src))
		n, err := encryptor.Transforming(dst, src, nil)
		if err != nil || n != len(src) {
			t.Fatalf("encrypt failed: n=%d err=%v", n, err)
		}

		plain := make([]byte, len(dst))
		n, err = decrypter.Transforming(plain, dst, nil)
		if err != nil || n != len(src) {
			t.Fatalf("decrypt failed: n=%d err=%v", n, err)
		}
		if !bytes.Equal(plain, src) {
			t.Fatal("unexpected decrypted payload")
		}
	})

	t.Run("aes-gcm", func(t *testing.T) {
		key := bytes.Repeat([]byte{3}, 16)
		encryptor, decrypter, err := NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, bytes.Repeat([]byte{4}, 16), nil)
		if err != nil {
			t.Fatalf("NewCipher failed: %v", err)
		}

		src := []byte("aead payload")
		nonce := bytes.Repeat([]byte{5}, encryptor.NonceSize())
		encrypted := make([]byte, encryptor.OutputSize(len(src)))
		if _, err := encryptor.Transforming(encrypted, src, nonce); err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		plain := make([]byte, decrypter.OutputSize(len(encrypted)))
		if _, err := decrypter.Transforming(plain, encrypted, nonce); err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if !bytes.Equal(plain, src) {
			t.Fatal("unexpected decrypted payload")
		}
	})
}

func TestCipherTransformingErrors(t *testing.T) {
	t.Run("aead dst too small", func(t *testing.T) {
		key := bytes.Repeat([]byte{6}, 16)
		encryptor, _, err := NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, bytes.Repeat([]byte{7}, 16), nil)
		if err != nil {
			t.Fatalf("NewCipher failed: %v", err)
		}
		if _, err := encryptor.Transforming(make([]byte, 1), []byte("hello"), bytes.Repeat([]byte{8}, encryptor.NonceSize())); err == nil {
			t.Fatal("expected dst too small error")
		}
	})

	t.Run("cipher stream panic converted to error", func(t *testing.T) {
		s := _CipherStream{}
		if _, err := s.Transforming(make([]byte, 1), []byte("x"), nil); err == nil {
			t.Fatal("expected transformed panic error")
		}
	})

	t.Run("block mode panic converted to error", func(t *testing.T) {
		s := _BlockModeEncryptor{BlockMode: cipher.NewCBCEncrypter(mustAESBlock(t), bytes.Repeat([]byte{1}, 16))}
		if _, err := s.Transforming(make([]byte, 1), []byte("x"), nil); err == nil {
			t.Fatal("expected transformed panic error")
		}
	})

	t.Run("aead decrypt tampered", func(t *testing.T) {
		key := bytes.Repeat([]byte{9}, 16)
		_, decrypter, err := NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, bytes.Repeat([]byte{10}, 16), nil)
		if err != nil {
			t.Fatalf("NewCipher failed: %v", err)
		}

		encrypted := bytes.Repeat([]byte{1}, decrypter.Overhead())
		dst := make([]byte, max(1, decrypter.OutputSize(len(encrypted))))
		if _, err := decrypter.Transforming(dst, encrypted, bytes.Repeat([]byte{2}, decrypter.NonceSize())); err == nil {
			t.Fatal("expected decrypt error")
		}
	})
}

func TestCipherSizeHelpers(t *testing.T) {
	block := _BlockModeEncryptor{BlockMode: cipher.NewCBCEncrypter(mustAESBlock(t), bytes.Repeat([]byte{1}, 16))}
	if !block.Pad() || block.Unpad() {
		t.Fatal("unexpected block encryptor pad flags")
	}
	if block.InputSize(16) != 32 || block.OutputSize(16) != 32 {
		t.Fatalf("unexpected block encryptor size helpers")
	}

	decrypter := _BlockModeDecrypter{BlockMode: cipher.NewCBCDecrypter(mustAESBlock(t), bytes.Repeat([]byte{1}, 16))}
	if decrypter.Pad() || !decrypter.Unpad() {
		t.Fatal("unexpected block decrypter pad flags")
	}
	if decrypter.OutputSize(32) != 16 {
		t.Fatalf("unexpected block decrypter output size: %d", decrypter.OutputSize(32))
	}

	stream := _CipherStream{}
	if stream.OutputSize(7) != 7 || stream.InputSize(7) != 7 {
		t.Fatal("unexpected stream size helpers")
	}
}

func mustAESBlock(t *testing.T) cipher.Block {
	t.Helper()
	block, err := NewCipherBlock(gtp.SymmetricEncryption_AES, bytes.Repeat([]byte{1}, 16))
	if err != nil {
		t.Fatalf("NewCipherBlock failed: %v", err)
	}
	return block
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

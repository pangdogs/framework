package method

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewSignerRSA(t *testing.T) {
	cases := []struct {
		name    string
		padding gtp.PaddingMode
		hash    gtp.Hash
	}{
		{name: "pkcs1v15-sha256", padding: gtp.PaddingMode_Pkcs1v15, hash: gtp.Hash_SHA256},
		{name: "pss-sha256", padding: gtp.PaddingMode_PSS, hash: gtp.Hash_SHA256},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			signer, err := NewSigner(gtp.AsymmetricEncryption_RSA, tc.padding, tc.hash)
			if err != nil {
				t.Fatalf("NewSigner failed: %v", err)
			}

			priv, err := signer.GenerateKey()
			if err != nil {
				t.Fatalf("GenerateKey failed: %v", err)
			}

			sig, err := signer.Sign(priv, []byte("rsa-message"))
			if err != nil {
				t.Fatalf("Sign failed: %v", err)
			}

			pub := priv.(*rsa.PrivateKey).Public()
			if err := signer.Verify(pub, []byte("rsa-message"), sig); err != nil {
				t.Fatalf("Verify failed: %v", err)
			}

			if err := signer.Verify(pub, []byte("tampered"), sig); err == nil {
				t.Fatal("expected verify error for tampered message")
			}
		})
	}
}

func TestNewSignerECDSA(t *testing.T) {
	signer, err := NewSigner(gtp.AsymmetricEncryption_ECDSA, 0, gtp.Hash_SHA256)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	priv, err := signer.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	sig, err := signer.Sign(priv, []byte("ecdsa-message"))
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	pub := priv.(*ecdsa.PrivateKey).Public()
	if err := signer.Verify(pub, []byte("ecdsa-message"), sig); err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
}

func TestNewSignerInvalid(t *testing.T) {
	t.Run("invalid method", func(t *testing.T) {
		if _, err := NewSigner(gtp.AsymmetricEncryption(255), 0, 0); err != ErrInvalidMethod {
			t.Fatalf("expected ErrInvalidMethod, got %v", err)
		}
	})

	t.Run("rsa invalid padding", func(t *testing.T) {
		if _, err := NewSigner(gtp.AsymmetricEncryption_RSA, gtp.PaddingMode_X923, gtp.Hash_SHA256); err == nil {
			t.Fatal("expected invalid padding error")
		}
	})

	t.Run("rsa invalid hash", func(t *testing.T) {
		if _, err := NewSigner(gtp.AsymmetricEncryption_RSA, gtp.PaddingMode_Pkcs1v15, gtp.Hash(255)); err == nil {
			t.Fatal("expected invalid hash error")
		}
	})
}

func TestSignerInvalidKeyTypes(t *testing.T) {
	rsaSigner, err := NewSigner(gtp.AsymmetricEncryption_RSA, gtp.PaddingMode_Pkcs1v15, gtp.Hash_SHA256)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	if _, err := rsaSigner.Sign(&ecdsa.PrivateKey{}, []byte("x")); err == nil {
		t.Fatal("expected RSA Sign type error")
	}
	if err := rsaSigner.Verify(&ecdsa.PublicKey{}, []byte("x"), []byte("sig")); err == nil {
		t.Fatal("expected RSA Verify type error")
	}

	ecdsaSigner, err := NewSigner(gtp.AsymmetricEncryption_ECDSA, 0, gtp.Hash_SHA256)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	if _, err := ecdsaSigner.Sign(&rsa.PrivateKey{}, []byte("x")); err == nil {
		t.Fatal("expected ECDSA Sign type error")
	}
}

func TestECDSAVerifyInvalidPublicKey(t *testing.T) {
	signer, err := NewSigner(gtp.AsymmetricEncryption_ECDSA, 0, gtp.Hash_SHA256)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	sig, err := signer.Sign(priv, []byte("hello"))
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	defer func() {
		if recover() != nil {
			t.Fatal("Verify should not panic on invalid public key type")
		}
	}()

	if err := signer.Verify(&rsa.PublicKey{}, []byte("hello"), sig); err == nil {
		t.Fatal("expected invalid public key error")
	}
}

func TestRSASignerGenerateKeyInvalidHash(t *testing.T) {
	signer := _RSASigner{}
	if _, err := signer.GenerateKey(); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestECDSASignerGenerateKeyInvalidHash(t *testing.T) {
	signer := _ECDSAPSigner{}
	if _, err := signer.GenerateKey(); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestECDSAVerifyTamperedSignature(t *testing.T) {
	signer, err := NewSigner(gtp.AsymmetricEncryption_ECDSA, 0, gtp.Hash_SHA256)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	priv, err := signer.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	sig, err := signer.Sign(priv, []byte("hello"))
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	sig[0] ^= 0xFF

	err = signer.Verify(priv.(*ecdsa.PrivateKey).Public(), []byte("hello"), sig)
	if err == nil || errors.Is(err, ErrInvalidMethod) {
		t.Fatalf("expected verification error, got %v", err)
	}
}

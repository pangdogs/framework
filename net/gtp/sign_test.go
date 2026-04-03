package gtp

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAndLoadRSAKeys(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})

	readPriv, err := ReadPrivateKey(bytes.NewReader(privPEM))
	if err != nil {
		t.Fatalf("ReadPrivateKey failed: %v", err)
	}
	if readPriv.N.Cmp(key.N) != 0 {
		t.Fatal("unexpected private key data")
	}

	readPub, err := ReadPublicKey(bytes.NewReader(pubPEM))
	if err != nil {
		t.Fatalf("ReadPublicKey failed: %v", err)
	}
	if readPub.N.Cmp(key.PublicKey.N) != 0 {
		t.Fatal("unexpected public key data")
	}

	tempDir := t.TempDir()
	privPath := filepath.Join(tempDir, "private.pem")
	pubPath := filepath.Join(tempDir, "public.pem")

	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		t.Fatalf("WriteFile private failed: %v", err)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0o600); err != nil {
		t.Fatalf("WriteFile public failed: %v", err)
	}

	filePriv, err := LoadPrivateKeyFile(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKeyFile failed: %v", err)
	}
	if filePriv.N.Cmp(key.N) != 0 {
		t.Fatal("unexpected loaded private key data")
	}

	filePub, err := LoadPublicKeyFile(pubPath)
	if err != nil {
		t.Fatalf("LoadPublicKeyFile failed: %v", err)
	}
	if filePub.N.Cmp(key.PublicKey.N) != 0 {
		t.Fatal("unexpected loaded public key data")
	}
}

func TestReadKeyInvalidPEM(t *testing.T) {
	if _, err := ReadPrivateKey(bytes.NewReader([]byte("not a pem"))); err == nil {
		t.Fatal("expected invalid private key error")
	}
	if _, err := ReadPublicKey(bytes.NewReader([]byte("not a pem"))); err == nil {
		t.Fatal("expected invalid public key error")
	}
}

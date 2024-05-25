package gtp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
)

// LoadPublicKeyFile 加载公钥文件
func LoadPublicKeyFile(filePath string) (*rsa.PublicKey, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadPublicKey(f)
}

// LoadPrivateKeyFile 加载私钥文件
func LoadPrivateKeyFile(filePath string) (*rsa.PrivateKey, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadPrivateKey(f)
}

// ReadPublicKey 读取公钥
func ReadPublicKey(reader io.Reader) (*rsa.PublicKey, error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bs)

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// ReadPrivateKey 读取私钥
func ReadPrivateKey(reader io.Reader) (*rsa.PrivateKey, error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bs)

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

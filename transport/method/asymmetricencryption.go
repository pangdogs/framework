package method

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"kit.golaxy.org/plugins/transport"
)

// Signer 签名器
type Signer interface {
	// GenerateKey 生成私钥
	GenerateKey() (crypto.PrivateKey, error)
	// Sign 签名
	Sign(priv crypto.PrivateKey, data []byte) ([]byte, error)
	// Verify 验证签名
	Verify(pub crypto.PublicKey, data, sig []byte) error
}

// NewSigner 创建签名器
func NewSigner(ae transport.AsymmetricEncryption, padding transport.PaddingMode, hash transport.Hash) (Signer, error) {
	switch ae {
	case transport.AsymmetricEncryption_RSA_256:
		switch padding {
		case transport.PaddingMode_Pkcs1v15, transport.PaddingMode_PSS:
			break
		default:
			return nil, errors.New("invalid padding mode")
		}

		var cryptoHash crypto.Hash
		switch hash {
		case transport.Hash_SHA256:
			cryptoHash = crypto.SHA256
		default:
			return nil, errors.New("invalid hash method")
		}

		return _RSA256Signer{padding: padding, hash: cryptoHash}, nil

	case transport.AsymmetricEncryption_ECDSA_P256:
		return _ECDSAP256Signer{}, nil
	default:
		return nil, ErrInvalidMethod
	}
}

type _RSA256Signer struct {
	padding transport.PaddingMode
	hash    crypto.Hash
}

func (s _RSA256Signer) GenerateKey() (crypto.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 256)
}

func (s _RSA256Signer) Sign(priv crypto.PrivateKey, data []byte) ([]byte, error) {
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid private key")
	}

	if !s.hash.Available() {
		return nil, errors.New("invalid hash method")
	}

	hashed := s.hash.New().Sum(data)

	switch s.padding {
	case transport.PaddingMode_Pkcs1v15:
		return rsa.SignPKCS1v15(rand.Reader, rsaPriv, s.hash, hashed)
	case transport.PaddingMode_PSS:
		return rsa.SignPSS(rand.Reader, rsaPriv, s.hash, hashed, nil)
	default:
		return nil, errors.New("invalid padding mode")
	}
}

func (s _RSA256Signer) Verify(pub crypto.PublicKey, hashed, sig []byte) error {
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("invalid public key")
	}

	switch s.padding {
	case transport.PaddingMode_Pkcs1v15:
		return rsa.VerifyPKCS1v15(rsaPub, s.hash, hashed, sig)
	case transport.PaddingMode_PSS:
		return rsa.VerifyPSS(rsaPub, s.hash, hashed, sig, nil)
	default:
		return errors.New("invalid padding mode")
	}
}

type _ECDSAP256Signer struct {
	hash crypto.Hash
}

func (s _ECDSAP256Signer) GenerateKey() (crypto.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func (s _ECDSAP256Signer) Sign(priv crypto.PrivateKey, data []byte) ([]byte, error) {
	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid private key")
	}

	hashed := s.hash.New().Sum(data)

	return ecdsa.SignASN1(rand.Reader, ecdsaPriv, hashed)
}

func (s _ECDSAP256Signer) Verify(pub crypto.PublicKey, hashed, sig []byte) error {
	if ecdsa.VerifyASN1(pub.(*ecdsa.PublicKey), hashed, sig) {
		return nil
	}
	return errors.New("crypto/ecdsa: verification error")
}

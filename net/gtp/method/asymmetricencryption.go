package method

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"git.golaxy.org/framework/net/gtp"
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
func NewSigner(ae gtp.AsymmetricEncryption, padding gtp.PaddingMode, hash gtp.Hash) (Signer, error) {
	switch ae {
	case gtp.AsymmetricEncryption_RSA256:
		switch padding {
		case gtp.PaddingMode_Pkcs1v15, gtp.PaddingMode_PSS:
			break
		default:
			return nil, errors.New("invalid padding mode")
		}

		var cryptoHash crypto.Hash
		switch hash {
		case gtp.Hash_SHA256:
			cryptoHash = crypto.SHA256
		default:
			return nil, errors.New("invalid hash method")
		}

		return _RSA256Signer{padding: padding, hash: cryptoHash}, nil

	case gtp.AsymmetricEncryption_ECDSA_P256:
		return _ECDSAP256Signer{}, nil
	default:
		return nil, ErrInvalidMethod
	}
}

type _RSA256Signer struct {
	padding gtp.PaddingMode
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

	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	switch s.padding {
	case gtp.PaddingMode_Pkcs1v15:
		return rsa.SignPKCS1v15(rand.Reader, rsaPriv, s.hash, hashed)
	case gtp.PaddingMode_PSS:
		return rsa.SignPSS(rand.Reader, rsaPriv, s.hash, hashed, nil)
	default:
		return nil, errors.New("invalid padding mode")
	}
}

func (s _RSA256Signer) Verify(pub crypto.PublicKey, data, sig []byte) error {
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("invalid public key")
	}

	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	switch s.padding {
	case gtp.PaddingMode_Pkcs1v15:
		return rsa.VerifyPKCS1v15(rsaPub, s.hash, hashed, sig)
	case gtp.PaddingMode_PSS:
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

	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	return ecdsa.SignASN1(rand.Reader, ecdsaPriv, hashed)
}

func (s _ECDSAP256Signer) Verify(pub crypto.PublicKey, data, sig []byte) error {
	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	if ecdsa.VerifyASN1(pub.(*ecdsa.PublicKey), hashed, sig) {
		return nil
	}

	return errors.New("crypto/ecdsa: verification error")
}

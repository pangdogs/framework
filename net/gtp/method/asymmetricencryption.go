/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

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
	case gtp.AsymmetricEncryption_RSA:
		switch padding {
		case gtp.PaddingMode_Pkcs1v15, gtp.PaddingMode_PSS:
			break
		default:
			return nil, errors.New("crypto/rsa: invalid padding mode")
		}

		cryptoHash, ok := toCryptoHash(hash)
		if !ok {
			return nil, errors.New("crypto/rsa: invalid hash method")
		}

		return _RSASigner{padding: padding, hash: cryptoHash}, nil

	case gtp.AsymmetricEncryption_ECDSA:
		cryptoHash, ok := toCryptoHash(hash)
		if !ok {
			return nil, errors.New("crypto/ecdsa: invalid hash method")
		}
		return _ECDSAPSigner{hash: cryptoHash}, nil

	default:
		return nil, ErrInvalidMethod
	}
}

type _RSASigner struct {
	padding gtp.PaddingMode
	hash    crypto.Hash
}

func (s _RSASigner) GenerateKey() (crypto.PrivateKey, error) {
	if !s.hash.Available() {
		return nil, errors.New("crypto/rsa: invalid hash method")
	}

	switch s.hash.Size() {
	case 32:
		return rsa.GenerateKey(rand.Reader, 2048)
	case 48:
		return rsa.GenerateKey(rand.Reader, 3072)
	case 64:
		return rsa.GenerateKey(rand.Reader, 4096)
	default:
		return nil, errors.New("crypto/rsa: invalid hash method")
	}
}

func (s _RSASigner) Sign(priv crypto.PrivateKey, data []byte) ([]byte, error) {
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("crypto/rsa: invalid private key")
	}

	if !s.hash.Available() {
		return nil, errors.New("crypto/rsa: invalid hash method")
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
		return nil, errors.New("crypto/rsa: invalid padding mode")
	}
}

func (s _RSASigner) Verify(pub crypto.PublicKey, data, sig []byte) error {
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("crypto/rsa: invalid public key")
	}
	if !s.hash.Available() {
		return errors.New("crypto/rsa: invalid hash method")
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
		return errors.New("crypto/rsa: invalid padding mode")
	}
}

type _ECDSAPSigner struct {
	hash crypto.Hash
}

func (s _ECDSAPSigner) GenerateKey() (crypto.PrivateKey, error) {
	if !s.hash.Available() {
		return nil, errors.New("crypto/ecdsa: invalid hash method")
	}

	switch s.hash.Size() {
	case 32:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case 48:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case 64:
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, errors.New("crypto/ecdsa: invalid hash method")
	}
}

func (s _ECDSAPSigner) Sign(priv crypto.PrivateKey, data []byte) ([]byte, error) {
	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("crypto/ecdsa: invalid private key")
	}
	if !s.hash.Available() {
		return nil, errors.New("crypto/ecdsa: invalid hash method")
	}

	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	return ecdsa.SignASN1(rand.Reader, ecdsaPriv, hashed)
}

func (s _ECDSAPSigner) Verify(pub crypto.PublicKey, data, sig []byte) error {
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("crypto/ecdsa: invalid public key")
	}
	if !s.hash.Available() {
		return errors.New("crypto/ecdsa: invalid hash method")
	}

	hash := s.hash.New()
	hash.Write(data)

	hashed := hash.Sum(nil)

	if ecdsa.VerifyASN1(ecdsaPub, hashed, sig) {
		return nil
	}

	return errors.New("crypto/ecdsa: verification error")
}

func toCryptoHash(hash gtp.Hash) (crypto.Hash, bool) {
	switch hash {
	case gtp.Hash_SHA256:
		return crypto.SHA256, true
	case gtp.Hash_SHA384:
		return crypto.SHA384, true
	case gtp.Hash_SHA512:
		return crypto.SHA512, true
	case gtp.Hash_BLAKE2b256:
		return crypto.BLAKE2b_256, true
	case gtp.Hash_BLAKE2b384:
		return crypto.BLAKE2b_384, true
	case gtp.Hash_BLAKE2b512:
		return crypto.BLAKE2b_512, true
	case gtp.Hash_BLAKE2s256:
		return crypto.BLAKE2s_256, true
	default:
		return 0, false
	}
}

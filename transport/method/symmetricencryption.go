package method

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"golang.org/x/crypto/chacha20"
	"kit.golaxy.org/plugins/transport"
)

// CipherStream 密码流
type CipherStream interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) error
	// Parallel 可否并行执行
	Parallel() bool
}

// NewCipherStream 创建密码流
func NewCipherStream(se transport.SymmetricEncryption, bcm transport.BlockCipherMode, key, iv []byte) (encrypter, decrypter CipherStream, err error) {
	switch se {
	case transport.SymmetricEncryption_AES:
		block, err := NewBlock(se, key)
		if err != nil {
			return nil, nil, err
		}
		return NewBlockCipherMode(bcm, block, iv)
	case transport.SymmetricEncryption_ChaCha20:
		encryptStream, err := chacha20.NewUnauthenticatedCipher(key, iv)
		if err != nil {
			return nil, nil, err
		}
		decryptStream, err := chacha20.NewUnauthenticatedCipher(key, iv)
		if err != nil {
			return nil, nil, err
		}
		encrypter = _XORKeyStream{Stream: encryptStream}
		decrypter = _XORKeyStream{Stream: decryptStream}
		return encrypter, decrypter, nil
	default:
		return nil, nil, ErrInvalidMethod
	}
}

// NewBlock 创建分组
func NewBlock(se transport.SymmetricEncryption, key []byte) (block cipher.Block, err error) {
	defer func() {
		if info := recover(); info != nil {
			panicErr, ok := info.(error)
			if ok {
				err = panicErr
			} else {
				err = fmt.Errorf("%v", info)
			}
		}
	}()

	switch se {
	case transport.SymmetricEncryption_AES:
		return aes.NewCipher(key)
	default:
		return nil, ErrInvalidMethod
	}
}

// NewBlockCipherMode 创建分组密码模式
func NewBlockCipherMode(bcm transport.BlockCipherMode, block cipher.Block, iv []byte) (encrypter, decrypter CipherStream, err error) {
	defer func() {
		if info := recover(); info != nil {
			panicErr, ok := info.(error)
			if ok {
				err = panicErr
			} else {
				err = fmt.Errorf("%v", info)
			}
		}
	}()

	switch bcm {
	case transport.BlockCipherMode_CTR:
		encrypter = _XORKeyStream{Stream: cipher.NewCTR(block, iv)}
		decrypter = _XORKeyStream{Stream: cipher.NewCTR(block, iv)}
		return
	case transport.BlockCipherMode_CFB:
		encrypter = _XORKeyStream{Stream: cipher.NewCFBEncrypter(block, iv)}
		decrypter = _XORKeyStream{Stream: cipher.NewCFBDecrypter(block, iv)}
		return
	case transport.BlockCipherMode_OFB:
		encrypter = _XORKeyStream{Stream: cipher.NewOFB(block, iv)}
		decrypter = _XORKeyStream{Stream: cipher.NewOFB(block, iv)}
		return
	case transport.BlockCipherMode_GCM:
		gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
		if err != nil {
			return nil, nil, err
		}
		encrypter = _AEADEncryptStream{AEAD: gcm, nonce: iv, parallel: true}
		decrypter = _AEADDecryptStream{AEAD: gcm, nonce: iv, parallel: true}
		return encrypter, decrypter, nil
	default:
		return nil, nil, ErrInvalidMethod
	}
}

type _XORKeyStream struct {
	cipher.Stream
	parallel bool
}

func (s _XORKeyStream) Transforming(dst, src []byte) error {
	s.XORKeyStream(dst, src)
	return nil
}

func (s _XORKeyStream) Parallel() bool {
	return s.parallel
}

type _AEADEncryptStream struct {
	cipher.AEAD
	nonce    []byte
	parallel bool
}

func (s _AEADEncryptStream) Transforming(dst, src []byte) error {
	s.Seal(dst, s.nonce, src, nil)
	return nil
}

func (s _AEADEncryptStream) Parallel() bool {
	return s.parallel
}

type _AEADDecryptStream struct {
	cipher.AEAD
	nonce    []byte
	parallel bool
}

func (s _AEADDecryptStream) Transforming(dst, src []byte) error {
	s.Open(dst, s.nonce, src, nil)
	return nil
}

func (s _AEADDecryptStream) Parallel() bool {
	return s.parallel
}

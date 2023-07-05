package method

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/transport"
)

// CipherStream 密码流
type CipherStream interface {
	// Transforming 变换数据
	Transforming(dst, src, nonce []byte) (int, error)
	// BlockSize block大小
	BlockSize() int
	// NonceSize nonce大小
	NonceSize() int
	// Overhead overhead大小
	Overhead() int
	// Pad 是否需要填充
	Pad() bool
	// Unpad 是否需要解除填充
	Unpad() bool
	// InputSize 输入大小
	InputSize(size int) int
	// OutputSize 输出大小
	OutputSize(size int) int
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

	case transport.SymmetricEncryption_ChaCha20, transport.SymmetricEncryption_XChaCha20:
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

	case transport.SymmetricEncryption_ChaCha20_Poly1305:
		encryptStream, err := chacha20poly1305.New(key)
		if err != nil {
			return nil, nil, err
		}
		decryptStream, err := chacha20poly1305.New(key)
		if err != nil {
			return nil, nil, err
		}
		encrypter = _AEADEncryptStream{AEAD: encryptStream}
		decrypter = _AEADDecryptStream{AEAD: decryptStream}
		return encrypter, decrypter, nil

	case transport.SymmetricEncryption_XChaCha20_Poly1305:
		encryptStream, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, nil, err
		}
		decryptStream, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, nil, err
		}
		encrypter = _AEADEncryptStream{AEAD: encryptStream}
		decrypter = _AEADDecryptStream{AEAD: decryptStream}
		return encrypter, decrypter, nil

	default:
		return nil, nil, ErrInvalidMethod
	}
}

// NewBlock 创建分组
func NewBlock(se transport.SymmetricEncryption, key []byte) (block cipher.Block, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			block = nil
			err = panicErr
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
		if panicErr := util.Panic2Err(); panicErr != nil {
			encrypter = nil
			decrypter = nil
			err = panicErr
		}
	}()

	switch bcm {
	case transport.BlockCipherMode_CTR:
		encrypter = _XORKeyStream{Stream: cipher.NewCTR(block, iv)}
		decrypter = _XORKeyStream{Stream: cipher.NewCTR(block, iv)}
		return
	case transport.BlockCipherMode_CBC:
		encrypter = _BlockModeEncryptStream{BlockMode: cipher.NewCBCEncrypter(block, iv)}
		decrypter = _BlockModeDecryptStream{BlockMode: cipher.NewCBCDecrypter(block, iv)}
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
		mode, err := cipher.NewGCM(block)
		if err != nil {
			return nil, nil, err
		}
		encrypter = _AEADEncryptStream{AEAD: mode}
		decrypter = _AEADDecryptStream{AEAD: mode}
		return encrypter, decrypter, nil
	default:
		return nil, nil, ErrInvalidMethod
	}
}

type _XORKeyStream struct {
	cipher.Stream
}

func (s _XORKeyStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			size = 0
			err = panicErr
		}
	}()
	s.XORKeyStream(dst, src)
	return len(dst), nil
}

func (s _XORKeyStream) BlockSize() int {
	return 0
}

func (s _XORKeyStream) NonceSize() int {
	return 0
}

func (s _XORKeyStream) Overhead() int {
	return 0
}

func (s _XORKeyStream) Pad() bool {
	return false
}

func (s _XORKeyStream) Unpad() bool {
	return false
}

func (s _XORKeyStream) InputSize(size int) int {
	return size
}

func (s _XORKeyStream) OutputSize(size int) int {
	return size
}

type _BlockModeEncryptStream struct {
	cipher.BlockMode
}

func (s _BlockModeEncryptStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			size = 0
			err = panicErr
		}
	}()
	s.CryptBlocks(dst, src)
	return len(dst), nil
}

func (s _BlockModeEncryptStream) NonceSize() int {
	return 0
}

func (s _BlockModeEncryptStream) Overhead() int {
	return 0
}

func (s _BlockModeEncryptStream) Pad() bool {
	return true
}

func (s _BlockModeEncryptStream) Unpad() bool {
	return false
}

func (s _BlockModeEncryptStream) InputSize(size int) int {
	return size + (s.BlockSize() - size%s.BlockSize())
}

func (s _BlockModeEncryptStream) OutputSize(size int) int {
	return size + (s.BlockSize() - size%s.BlockSize())
}

type _BlockModeDecryptStream struct {
	cipher.BlockMode
}

func (s _BlockModeDecryptStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			size = 0
			err = panicErr
		}
	}()
	s.CryptBlocks(dst, src)
	return len(dst), nil
}

func (s _BlockModeDecryptStream) NonceSize() int {
	return 0
}

func (s _BlockModeDecryptStream) Overhead() int {
	return 0
}

func (s _BlockModeDecryptStream) Pad() bool {
	return false
}

func (s _BlockModeDecryptStream) Unpad() bool {
	return true
}

func (s _BlockModeDecryptStream) InputSize(size int) int {
	return size
}

func (s _BlockModeDecryptStream) OutputSize(size int) int {
	return size - (s.BlockSize() - size%s.BlockSize())
}

type _AEADEncryptStream struct {
	cipher.AEAD
}

func (s _AEADEncryptStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			size = 0
			err = panicErr
		}
	}()
	if len(dst) < s.OutputSize(len(src)) {
		return 0, errors.New("dst too small")
	}
	out := s.Seal(dst[:0], nonce, src, nil)
	return len(out), nil
}

func (s _AEADEncryptStream) BlockSize() int {
	return 0
}

func (s _AEADEncryptStream) Pad() bool {
	return false
}

func (s _AEADEncryptStream) Unpad() bool {
	return false
}

func (s _AEADEncryptStream) InputSize(size int) int {
	return size
}

func (s _AEADEncryptStream) OutputSize(size int) int {
	return size + s.Overhead()
}

type _AEADDecryptStream struct {
	cipher.AEAD
}

func (s _AEADDecryptStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			size = 0
			err = panicErr
		}
	}()
	if len(dst) < s.OutputSize(len(src)) {
		return 0, errors.New("dst too small")
	}
	var out []byte
	out, err = s.Open(dst[:0], nonce, src, nil)
	return len(out), err
}

func (s _AEADDecryptStream) BlockSize() int {
	return 0
}

func (s _AEADDecryptStream) Pad() bool {
	return false
}

func (s _AEADDecryptStream) Unpad() bool {
	return false
}

func (s _AEADDecryptStream) InputSize(size int) int {
	return size
}

func (s _AEADDecryptStream) OutputSize(size int) int {
	return size - s.Overhead()
}

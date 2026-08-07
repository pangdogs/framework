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
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/framework/net/gtp"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
)

// Cipher 是有状态的对称加密或解密变换，不支持并发调用。
type Cipher interface {
	// Transforming 使用 nonce 将 src 变换到 dst，并返回输出字节数。
	Transforming(dst, src, nonce []byte) (int, error)
	// BlockSize 返回分组密码块大小；非分组变换返回零。
	BlockSize() int
	// NonceSize 返回每次变换要求的 nonce 字节数。
	NonceSize() int
	// Overhead 返回认证加密相对明文增加的字节数。
	Overhead() int
	// Pad 报告变换前是否需要填充输入。
	Pad() bool
	// Unpad 报告变换后是否需要去除填充。
	Unpad() bool
	// InputSize 返回容纳 size 字节输入所需的缓冲区大小。
	InputSize(size int) int
	// OutputSize 返回 size 字节输入对应的输出缓冲区大小。
	OutputSize(size int) int
}

// NewCipher 根据算法和模式创建相互独立的加密器与解密器。
func NewCipher(se gtp.SymmetricEncryption, bcm gtp.BlockCipherMode, key, iv, nonce []byte) (encryptor, decrypter Cipher, err error) {
	switch se {
	case gtp.SymmetricEncryption_AES:
		block, err := NewCipherBlock(se, key)
		if err != nil {
			return nil, nil, err
		}
		return NewBlockCipherMode(bcm, block, iv)

	case gtp.SymmetricEncryption_ChaCha20, gtp.SymmetricEncryption_XChaCha20:
		_encryptor, err := chacha20.NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			return nil, nil, err
		}
		_decrypter, err := chacha20.NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			return nil, nil, err
		}
		encryptor = _CipherStream{Stream: _encryptor}
		decrypter = _CipherStream{Stream: _decrypter}
		return encryptor, decrypter, nil

	case gtp.SymmetricEncryption_ChaCha20_Poly1305:
		_encryptor, err := chacha20poly1305.New(key)
		if err != nil {
			return nil, nil, err
		}
		_decrypter, err := chacha20poly1305.New(key)
		if err != nil {
			return nil, nil, err
		}
		encryptor = _AEADEncryptor{AEAD: _encryptor}
		decrypter = _AEADDecrypter{AEAD: _decrypter}
		return encryptor, decrypter, nil

	case gtp.SymmetricEncryption_XChaCha20_Poly1305:
		_encryptor, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, nil, err
		}
		_decrypter, err := chacha20poly1305.NewX(key)
		if err != nil {
			return nil, nil, err
		}
		encryptor = _AEADEncryptor{AEAD: _encryptor}
		decrypter = _AEADDecrypter{AEAD: _decrypter}
		return encryptor, decrypter, nil

	default:
		return nil, nil, ErrInvalidMethod
	}
}

// NewCipherBlock 创建底层分组密码，并将构造 panic 转换为错误。
func NewCipherBlock(se gtp.SymmetricEncryption, key []byte) (block cipher.Block, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			block = nil
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	switch se {
	case gtp.SymmetricEncryption_AES:
		return aes.NewCipher(key)
	default:
		return nil, ErrInvalidMethod
	}
}

// NewBlockCipherMode 为分组密码创建成对的加密和解密变换，并将构造 panic 转换为错误。
func NewBlockCipherMode(bcm gtp.BlockCipherMode, block cipher.Block, iv []byte) (encryptor, decrypter Cipher, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			encryptor = nil
			decrypter = nil
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()

	switch bcm {
	case gtp.BlockCipherMode_CTR:
		encryptor = _CipherStream{Stream: cipher.NewCTR(block, iv)}
		decrypter = _CipherStream{Stream: cipher.NewCTR(block, iv)}
		return
	case gtp.BlockCipherMode_CBC:
		encryptor = _BlockModeEncryptor{BlockMode: cipher.NewCBCEncrypter(block, iv)}
		decrypter = _BlockModeDecrypter{BlockMode: cipher.NewCBCDecrypter(block, iv)}
		return
	case gtp.BlockCipherMode_CFB:
		encryptor = _CipherStream{Stream: cipher.NewCFBEncrypter(block, iv)}
		decrypter = _CipherStream{Stream: cipher.NewCFBDecrypter(block, iv)}
		return
	case gtp.BlockCipherMode_OFB:
		encryptor = _CipherStream{Stream: cipher.NewOFB(block, iv)}
		decrypter = _CipherStream{Stream: cipher.NewOFB(block, iv)}
		return
	case gtp.BlockCipherMode_GCM:
		mode, err := cipher.NewGCMWithNonceSize(block, block.BlockSize())
		if err != nil {
			return nil, nil, err
		}
		encryptor = _AEADEncryptor{AEAD: mode}
		decrypter = _AEADDecrypter{AEAD: mode}
		return encryptor, decrypter, nil
	default:
		return nil, nil, ErrInvalidMethod
	}
}

type _CipherStream struct {
	cipher.Stream
}

func (s _CipherStream) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			size = 0
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()
	s.XORKeyStream(dst, src)
	return len(dst), nil
}

func (s _CipherStream) BlockSize() int {
	return 0
}

func (s _CipherStream) NonceSize() int {
	return 0
}

func (s _CipherStream) Overhead() int {
	return 0
}

func (s _CipherStream) Pad() bool {
	return false
}

func (s _CipherStream) Unpad() bool {
	return false
}

func (s _CipherStream) InputSize(size int) int {
	return size
}

func (s _CipherStream) OutputSize(size int) int {
	return size
}

type _BlockModeEncryptor struct {
	cipher.BlockMode
}

func (s _BlockModeEncryptor) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			size = 0
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()
	s.CryptBlocks(dst, src)
	return len(dst), nil
}

func (s _BlockModeEncryptor) NonceSize() int {
	return 0
}

func (s _BlockModeEncryptor) Overhead() int {
	return 0
}

func (s _BlockModeEncryptor) Pad() bool {
	return true
}

func (s _BlockModeEncryptor) Unpad() bool {
	return false
}

func (s _BlockModeEncryptor) InputSize(size int) int {
	return size + (s.BlockSize() - size%s.BlockSize())
}

func (s _BlockModeEncryptor) OutputSize(size int) int {
	return size + (s.BlockSize() - size%s.BlockSize())
}

type _BlockModeDecrypter struct {
	cipher.BlockMode
}

func (s _BlockModeDecrypter) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			size = 0
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()
	s.CryptBlocks(dst, src)
	return len(dst), nil
}

func (s _BlockModeDecrypter) NonceSize() int {
	return 0
}

func (s _BlockModeDecrypter) Overhead() int {
	return 0
}

func (s _BlockModeDecrypter) Pad() bool {
	return false
}

func (s _BlockModeDecrypter) Unpad() bool {
	return true
}

func (s _BlockModeDecrypter) InputSize(size int) int {
	return size
}

func (s _BlockModeDecrypter) OutputSize(size int) int {
	return size - (s.BlockSize() - size%s.BlockSize())
}

type _AEADEncryptor struct {
	cipher.AEAD
}

func (s _AEADEncryptor) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			size = 0
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()
	if len(dst) < s.OutputSize(len(src)) {
		return 0, errors.New("crypto/aead: dst too small")
	}
	out := s.Seal(dst[:0], nonce, src, nil)
	return len(out), nil
}

func (s _AEADEncryptor) BlockSize() int {
	return 0
}

func (s _AEADEncryptor) Pad() bool {
	return false
}

func (s _AEADEncryptor) Unpad() bool {
	return false
}

func (s _AEADEncryptor) InputSize(size int) int {
	return size
}

func (s _AEADEncryptor) OutputSize(size int) int {
	return size + s.Overhead()
}

type _AEADDecrypter struct {
	cipher.AEAD
}

func (s _AEADDecrypter) Transforming(dst, src, nonce []byte) (size int, err error) {
	defer func() {
		if panicErr := types.Panic2Err(recover()); panicErr != nil {
			size = 0
			err = fmt.Errorf("%w: %w", core.ErrPanicked, panicErr)
		}
	}()
	if len(dst) < s.OutputSize(len(src)) {
		return 0, errors.New("crypto/aead: dst too small")
	}
	var out []byte
	out, err = s.Open(dst[:0], nonce, src, nil)
	return len(out), err
}

func (s _AEADDecrypter) BlockSize() int {
	return 0
}

func (s _AEADDecrypter) Pad() bool {
	return false
}

func (s _AEADDecrypter) Unpad() bool {
	return false
}

func (s _AEADDecrypter) InputSize(size int) int {
	return size
}

func (s _AEADDecrypter) OutputSize(size int) int {
	return size - s.Overhead()
}

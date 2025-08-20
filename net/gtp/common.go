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

package gtp

import (
	"crypto/aes"
	"fmt"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
	"strings"
)

// Version 协议版本
type Version uint16

const (
	Version_V1_0 Version = 0x0100 // 协议v1.0版本
)

// SecretKeyExchange 秘钥交换函数
type SecretKeyExchange uint8

const (
	SecretKeyExchange_None  SecretKeyExchange = iota // 未设置
	SecretKeyExchange_ECDHE                          // ECDHE算法
)

// ParseSecretKeyExchange 解析配置字串
func ParseSecretKeyExchange(str string) (SecretKeyExchange, error) {
	switch strings.ToLower(str) {
	case "none":
		return SecretKeyExchange_None, nil
	case "ecdhe":
		return SecretKeyExchange_ECDHE, nil
	default:
		return SecretKeyExchange_None, fmt.Errorf("%w: invalid SecretKeyExchange", ErrGTP)
	}
}

// String implements fmt.Stringer
func (ske SecretKeyExchange) String() string {
	switch ske {
	case SecretKeyExchange_ECDHE:
		return "ecdhe"
	default:
		return "none"
	}
}

// AsymmetricEncryption 非对称加密算法
type AsymmetricEncryption uint8

const (
	AsymmetricEncryption_None       AsymmetricEncryption = iota // 未设置
	AsymmetricEncryption_RSA256                                 // RSA-256算法
	AsymmetricEncryption_ECDSA_P256                             // ECDSA-NIST-P256算法
)

// ParseAsymmetricEncryption 解析配置字串
func ParseAsymmetricEncryption(str string) (AsymmetricEncryption, error) {
	switch strings.ToLower(str) {
	case "none":
		return AsymmetricEncryption_None, nil
	case "rsa256":
		return AsymmetricEncryption_RSA256, nil
	case "ecdsa_p256":
		return AsymmetricEncryption_ECDSA_P256, nil
	default:
		return AsymmetricEncryption_None, fmt.Errorf("%w: invalid AsymmetricEncryption", ErrGTP)
	}
}

// String implements fmt.Stringer
func (ae AsymmetricEncryption) String() string {
	switch ae {
	case AsymmetricEncryption_RSA256:
		return "rsa256"
	case AsymmetricEncryption_ECDSA_P256:
		return "ecdsa_p256"
	default:
		return "none"
	}
}

// SymmetricEncryption 对称加密算法
type SymmetricEncryption uint8

const (
	SymmetricEncryption_None               SymmetricEncryption = iota // 未设置
	SymmetricEncryption_AES                                           // AES算法（分组密码模式）
	SymmetricEncryption_ChaCha20                                      // ChaCha20算法（流模式）
	SymmetricEncryption_XChaCha20                                     // XChaCha20算法（流模式）
	SymmetricEncryption_ChaCha20_Poly1305                             // ChaCha20-Poly1305算法（流模式）
	SymmetricEncryption_XChaCha20_Poly1305                            // XChaCha20-Poly1305算法（流模式）
)

// ParseSymmetricEncryption 解析配置字串
func ParseSymmetricEncryption(str string) (SymmetricEncryption, error) {
	switch strings.ToLower(str) {
	case "none":
		return SymmetricEncryption_None, nil
	case "aes":
		return SymmetricEncryption_AES, nil
	case "chacha20":
		return SymmetricEncryption_ChaCha20, nil
	case "xchacha20":
		return SymmetricEncryption_XChaCha20, nil
	case "chacha20_poly1305":
		return SymmetricEncryption_ChaCha20_Poly1305, nil
	case "xchacha20_poly1305":
		return SymmetricEncryption_XChaCha20_Poly1305, nil
	default:
		return SymmetricEncryption_None, fmt.Errorf("%w: invalid SymmetricEncryption", ErrGTP)
	}
}

// String implements fmt.Stringer
func (se SymmetricEncryption) String() string {
	switch se {
	case SymmetricEncryption_AES:
		return "aes"
	case SymmetricEncryption_ChaCha20:
		return "chacha20"
	case SymmetricEncryption_XChaCha20:
		return "xchacha20"
	case SymmetricEncryption_ChaCha20_Poly1305:
		return "chacha20_poly1305"
	case SymmetricEncryption_XChaCha20_Poly1305:
		return "xchacha20_poly1305"
	default:
		return "none"
	}
}

// BlockSize 获取block大小，分组密码模式使用
func (se SymmetricEncryption) BlockSize() (int, bool) {
	switch se {
	case SymmetricEncryption_AES:
		return aes.BlockSize, true
	default:
		return 0, false
	}
}

// Nonce 获取nonce大小，流密码模式使用
func (se SymmetricEncryption) Nonce() (int, bool) {
	switch se {
	case SymmetricEncryption_ChaCha20:
		return chacha20.NonceSize, true
	case SymmetricEncryption_XChaCha20:
		return chacha20.NonceSizeX, true
	case SymmetricEncryption_ChaCha20_Poly1305:
		return chacha20poly1305.NonceSize, true
	case SymmetricEncryption_XChaCha20_Poly1305:
		return chacha20poly1305.NonceSizeX, true
	default:
		return 0, false
	}
}

// BlockCipherMode 是否需要使用分组密码模式
func (se SymmetricEncryption) BlockCipherMode() bool {
	switch se {
	case SymmetricEncryption_AES:
		return true
	default:
		return false
	}
}

// StreamCipherMode 是否需要使用流密码模式
func (se SymmetricEncryption) StreamCipherMode() bool {
	switch se {
	case SymmetricEncryption_ChaCha20, SymmetricEncryption_XChaCha20, SymmetricEncryption_ChaCha20_Poly1305, SymmetricEncryption_XChaCha20_Poly1305:
		return true
	default:
		return false
	}
}

// PaddingMode 数据填充方案
type PaddingMode uint8

const (
	PaddingMode_None     PaddingMode = iota // 未设置
	PaddingMode_Pkcs7                       // pkcs7方案（用于对称加密）
	PaddingMode_X923                        // x927方案（用于对称加密）
	PaddingMode_Pkcs1v15                    // Pkcs1v15方案（用于非对称加密RSA算法）
	PaddingMode_PSS                         // PSS方案（用于非对称加密RSA算法）
)

// ParsePaddingMode 解析配置字串
func ParsePaddingMode(str string) (PaddingMode, error) {
	switch strings.ToLower(str) {
	case "none":
		return PaddingMode_None, nil
	case "pkcs7":
		return PaddingMode_Pkcs7, nil
	case "x923":
		return PaddingMode_X923, nil
	case "pkcs1v15":
		return PaddingMode_Pkcs1v15, nil
	case "pss":
		return PaddingMode_PSS, nil
	default:
		return PaddingMode_None, fmt.Errorf("%w: invalid PaddingMode", ErrGTP)
	}
}

// String implements fmt.Stringer
func (pm PaddingMode) String() string {
	switch pm {
	case PaddingMode_Pkcs7:
		return "pkcs7"
	case PaddingMode_X923:
		return "x923"
	case PaddingMode_Pkcs1v15:
		return "pkcs1v15"
	case PaddingMode_PSS:
		return "pss"
	default:
		return "none"
	}
}

// BlockCipherMode 分组密码模式
type BlockCipherMode uint8

const (
	BlockCipherMode_None BlockCipherMode = iota // 未设置
	BlockCipherMode_CTR                         // CTR模式
	BlockCipherMode_CBC                         // CBC模式
	BlockCipherMode_CFB                         // CFB模式
	BlockCipherMode_OFB                         // OFB模式
	BlockCipherMode_GCM                         // GCM模式
)

// ParseBlockCipherMode 解析配置字串
func ParseBlockCipherMode(str string) (BlockCipherMode, error) {
	switch strings.ToLower(str) {
	case "none":
		return BlockCipherMode_None, nil
	case "ctr":
		return BlockCipherMode_CTR, nil
	case "cbc":
		return BlockCipherMode_CBC, nil
	case "cfb":
		return BlockCipherMode_CFB, nil
	case "ofb":
		return BlockCipherMode_OFB, nil
	case "gcm":
		return BlockCipherMode_GCM, nil
	default:
		return BlockCipherMode_None, fmt.Errorf("%w: invalid BlockCipherMode", ErrGTP)
	}
}

// String implements fmt.Stringer
func (bcm BlockCipherMode) String() string {
	switch bcm {
	case BlockCipherMode_CTR:
		return "ctr"
	case BlockCipherMode_CBC:
		return "cbc"
	case BlockCipherMode_CFB:
		return "cfb"
	case BlockCipherMode_OFB:
		return "ofb"
	case BlockCipherMode_GCM:
		return "gcm"
	default:
		return "none"
	}
}

// IV 是否需要iv，iv大小与加密算法的blocksize相同
func (bcm BlockCipherMode) IV() bool {
	switch bcm {
	case BlockCipherMode_CTR, BlockCipherMode_CBC, BlockCipherMode_CFB, BlockCipherMode_OFB:
		return true
	default:
		return false
	}
}

// Nonce 是否需要nonce，nonce大小与加密算法的blocksize相同
func (bcm BlockCipherMode) Nonce() bool {
	switch bcm {
	case BlockCipherMode_GCM:
		return true
	default:
		return false
	}
}

// Padding 是否需要分组对齐填充数据
func (bcm BlockCipherMode) Padding() bool {
	switch bcm {
	case BlockCipherMode_CBC:
		return true
	default:
		return false
	}
}

// Hash 摘要函数
type Hash uint8

const (
	Hash_None    Hash = iota // 未设置
	Hash_SHA256              // SHA256算法
	Hash_BLAKE2s             // BLAKE2b算法
)

// ParseHash 解析配置字串
func ParseHash(str string) (Hash, error) {
	switch strings.ToLower(str) {
	case "none":
		return Hash_None, nil
	case "sha256":
		return Hash_SHA256, nil
	case "blake2s":
		return Hash_BLAKE2s, nil
	default:
		return Hash_None, fmt.Errorf("%w: invalid Hash", ErrGTP)
	}
}

// String implements fmt.Stringer
func (h Hash) String() string {
	switch h {
	case Hash_SHA256:
		return "sha256"
	case Hash_BLAKE2s:
		return "blake2s"
	default:
		return "none"
	}
}

// Bits 位数
func (h Hash) Bits() int {
	switch h {
	case Hash_SHA256, Hash_BLAKE2s:
		return 256
	default:
		return 0
	}
}

// NamedCurve 曲线类型
type NamedCurve uint8

const (
	NamedCurve_None   NamedCurve = iota // 未设置
	NamedCurve_X25519                   // 曲线x25519
	NamedCurve_P256                     // 曲线NIST-P256
)

// ParseNamedCurve 解析配置字串
func ParseNamedCurve(str string) (NamedCurve, error) {
	switch strings.ToLower(str) {
	case "none":
		return NamedCurve_None, nil
	case "x25519":
		return NamedCurve_X25519, nil
	case "p256":
		return NamedCurve_P256, nil
	default:
		return NamedCurve_None, fmt.Errorf("%w: invalid NamedCurve", ErrGTP)
	}
}

// String implements fmt.Stringer
func (nc NamedCurve) String() string {
	switch nc {
	case NamedCurve_X25519:
		return "x25519"
	case NamedCurve_P256:
		return "p256"
	default:
		return "none"
	}
}

// Compression 压缩函数
type Compression uint8

const (
	Compression_None    Compression = iota // 未设置
	Compression_Gzip                       // Gzip压缩算法
	Compression_Deflate                    // Deflate压缩算法
	Compression_Brotli                     // Brotli压缩算法
	Compression_LZ4                        // LZ4压缩算法
	Compression_Snappy                     // Snappy压缩算法
)

// ParseCompression 解析配置字串
func ParseCompression(str string) (Compression, error) {
	switch strings.ToLower(str) {
	case "none":
		return Compression_None, nil
	case "gzip":
		return Compression_Gzip, nil
	case "deflate":
		return Compression_Deflate, nil
	case "brotli":
		return Compression_Brotli, nil
	case "lz4":
		return Compression_LZ4, nil
	case "snappy":
		return Compression_Snappy, nil
	default:
		return Compression_None, fmt.Errorf("%w: invalid Compression", ErrGTP)
	}
}

// String implements fmt.Stringer
func (c Compression) String() string {
	switch c {
	case Compression_Gzip:
		return "gzip"
	case Compression_Deflate:
		return "deflate"
	case Compression_Brotli:
		return "brotli"
	case Compression_LZ4:
		return "lz4"
	case Compression_Snappy:
		return "snappy"
	default:
		return "none"
	}
}

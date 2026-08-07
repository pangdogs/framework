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
	"strings"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
)

// Version 表示 GTP 协议版本。
type Version uint16

const (
	// Version_V1_0 表示 GTP 1.0。
	Version_V1_0 Version = 0x0100
)

// SecretKeyExchange 标识会话密钥交换算法。
type SecretKeyExchange uint8

const (
	// SecretKeyExchange_None 表示不交换会话密钥。
	SecretKeyExchange_None SecretKeyExchange = iota
	// SecretKeyExchange_ECDHE 表示使用临时椭圆曲线 Diffie-Hellman 密钥交换。
	SecretKeyExchange_ECDHE
)

// ParseSecretKeyExchange 解析不区分大小写的密钥交换算法名称。
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

// String 返回用于配置的密钥交换算法名称。
func (ske SecretKeyExchange) String() string {
	switch ske {
	case SecretKeyExchange_ECDHE:
		return "ecdhe"
	default:
		return "none"
	}
}

// AsymmetricEncryption 标识握手签名使用的非对称算法。
type AsymmetricEncryption uint8

const (
	// AsymmetricEncryption_None 表示不使用非对称签名。
	AsymmetricEncryption_None AsymmetricEncryption = iota
	// AsymmetricEncryption_RSA 表示 RSA。
	AsymmetricEncryption_RSA
	// AsymmetricEncryption_ECDSA 表示基于 NIST 曲线的 ECDSA。
	AsymmetricEncryption_ECDSA
)

// ParseAsymmetricEncryption 解析不区分大小写的非对称算法名称。
func ParseAsymmetricEncryption(str string) (AsymmetricEncryption, error) {
	switch strings.ToLower(str) {
	case "none":
		return AsymmetricEncryption_None, nil
	case "rsa":
		return AsymmetricEncryption_RSA, nil
	case "ecdsa":
		return AsymmetricEncryption_ECDSA, nil
	default:
		return AsymmetricEncryption_None, fmt.Errorf("%w: invalid AsymmetricEncryption", ErrGTP)
	}
}

// String 返回用于配置的非对称算法名称。
func (ae AsymmetricEncryption) String() string {
	switch ae {
	case AsymmetricEncryption_RSA:
		return "rsa"
	case AsymmetricEncryption_ECDSA:
		return "ecdsa"
	default:
		return "none"
	}
}

// SymmetricEncryption 标识负载加密使用的对称算法。
type SymmetricEncryption uint8

const (
	// SymmetricEncryption_None 表示不加密负载。
	SymmetricEncryption_None SymmetricEncryption = iota
	// SymmetricEncryption_AES 表示 AES 分组密码。
	SymmetricEncryption_AES
	// SymmetricEncryption_ChaCha20 表示 ChaCha20 流密码。
	SymmetricEncryption_ChaCha20
	// SymmetricEncryption_XChaCha20 表示 XChaCha20 流密码。
	SymmetricEncryption_XChaCha20
	// SymmetricEncryption_ChaCha20_Poly1305 表示 ChaCha20-Poly1305 AEAD。
	SymmetricEncryption_ChaCha20_Poly1305
	// SymmetricEncryption_XChaCha20_Poly1305 表示 XChaCha20-Poly1305 AEAD。
	SymmetricEncryption_XChaCha20_Poly1305
)

// ParseSymmetricEncryption 解析不区分大小写的对称算法名称。
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

// String 返回用于配置的对称算法名称。
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

// BlockSize 返回分组密码的块大小；算法不是分组密码时返回 false。
func (se SymmetricEncryption) BlockSize() (int, bool) {
	switch se {
	case SymmetricEncryption_AES:
		return aes.BlockSize, true
	default:
		return 0, false
	}
}

// Nonce 返回流密码或 AEAD 的 nonce 字节数；算法不需要 nonce 时返回 false。
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

// BlockCipherMode 报告算法是否需要配合分组密码模式。
func (se SymmetricEncryption) BlockCipherMode() bool {
	switch se {
	case SymmetricEncryption_AES:
		return true
	default:
		return false
	}
}

// StreamCipherMode 报告算法是否属于流密码或 AEAD。
func (se SymmetricEncryption) StreamCipherMode() bool {
	switch se {
	case SymmetricEncryption_ChaCha20, SymmetricEncryption_XChaCha20, SymmetricEncryption_ChaCha20_Poly1305, SymmetricEncryption_XChaCha20_Poly1305:
		return true
	default:
		return false
	}
}

// PaddingMode 标识对称加密填充或 RSA 签名填充方案。
type PaddingMode uint8

const (
	// PaddingMode_None 表示不使用填充。
	PaddingMode_None PaddingMode = iota
	// PaddingMode_Pkcs7 表示对称加密使用 PKCS#7 填充。
	PaddingMode_Pkcs7
	// PaddingMode_X923 表示对称加密使用 ANSI X9.23 填充。
	PaddingMode_X923
	// PaddingMode_Pkcs1v15 表示 RSA 使用 PKCS#1 v1.5 签名填充。
	PaddingMode_Pkcs1v15
	// PaddingMode_PSS 表示 RSA 使用 PSS 签名填充。
	PaddingMode_PSS
)

// ParsePaddingMode 解析不区分大小写的填充方案名称。
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

// String 返回用于配置的填充方案名称。
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

// BlockCipherMode 标识 AES 等分组密码的工作模式。
type BlockCipherMode uint8

const (
	// BlockCipherMode_None 表示未设置分组密码模式。
	BlockCipherMode_None BlockCipherMode = iota
	// BlockCipherMode_CTR 表示 CTR 模式。
	BlockCipherMode_CTR
	// BlockCipherMode_CBC 表示 CBC 模式。
	BlockCipherMode_CBC
	// BlockCipherMode_CFB 表示 CFB 模式。
	BlockCipherMode_CFB
	// BlockCipherMode_OFB 表示 OFB 模式。
	BlockCipherMode_OFB
	// BlockCipherMode_GCM 表示 GCM AEAD 模式。
	BlockCipherMode_GCM
)

// ParseBlockCipherMode 解析不区分大小写的分组密码模式名称。
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

// String 返回用于配置的分组密码模式名称。
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

// IV 报告模式是否需要与密码块等长的初始化向量。
func (bcm BlockCipherMode) IV() bool {
	switch bcm {
	case BlockCipherMode_CTR, BlockCipherMode_CBC, BlockCipherMode_CFB, BlockCipherMode_OFB:
		return true
	default:
		return false
	}
}

// Nonce 报告模式是否需要 nonce。
func (bcm BlockCipherMode) Nonce() bool {
	switch bcm {
	case BlockCipherMode_GCM:
		return true
	default:
		return false
	}
}

// Padding 报告模式是否要求负载按块大小填充。
func (bcm BlockCipherMode) Padding() bool {
	switch bcm {
	case BlockCipherMode_CBC:
		return true
	default:
		return false
	}
}

// Hash 标识摘要算法。
type Hash uint8

const (
	// Hash_None 表示不使用摘要算法。
	Hash_None Hash = iota
	// Hash_SHA256 表示 SHA-256。
	Hash_SHA256
	// Hash_SHA384 表示 SHA-384。
	Hash_SHA384
	// Hash_SHA512 表示 SHA-512。
	Hash_SHA512
	// Hash_BLAKE2b256 表示 BLAKE2b-256。
	Hash_BLAKE2b256
	// Hash_BLAKE2b384 表示 BLAKE2b-384。
	Hash_BLAKE2b384
	// Hash_BLAKE2b512 表示 BLAKE2b-512。
	Hash_BLAKE2b512
	// Hash_BLAKE2s256 表示 BLAKE2s-256。
	Hash_BLAKE2s256
)

// ParseHash 解析不区分大小写的摘要算法名称。
func ParseHash(str string) (Hash, error) {
	switch strings.ToLower(str) {
	case "none":
		return Hash_None, nil
	case "sha256":
		return Hash_SHA256, nil
	case "sha384":
		return Hash_SHA384, nil
	case "sha512":
		return Hash_SHA512, nil
	case "blake2b256":
		return Hash_BLAKE2b256, nil
	case "blake2b384":
		return Hash_BLAKE2b384, nil
	case "blake2b512":
		return Hash_BLAKE2b512, nil
	case "blake2s256":
		return Hash_BLAKE2s256, nil
	default:
		return Hash_None, fmt.Errorf("%w: invalid Hash", ErrGTP)
	}
}

// String 返回用于配置的摘要算法名称。
func (h Hash) String() string {
	switch h {
	case Hash_SHA256:
		return "sha256"
	case Hash_SHA384:
		return "sha384"
	case Hash_SHA512:
		return "sha512"
	case Hash_BLAKE2b256:
		return "blake2b256"
	case Hash_BLAKE2b384:
		return "blake2b384"
	case Hash_BLAKE2b512:
		return "blake2b512"
	case Hash_BLAKE2s256:
		return "blake2s256"
	default:
		return "none"
	}
}

// NamedCurve 标识 ECDHE 或 ECDSA 使用的命名曲线。
type NamedCurve uint8

const (
	// NamedCurve_None 表示未设置命名曲线。
	NamedCurve_None NamedCurve = iota
	// NamedCurve_X25519 表示 X25519。
	NamedCurve_X25519
	// NamedCurve_P256 表示 NIST P-256。
	NamedCurve_P256
	// NamedCurve_P384 表示 NIST P-384。
	NamedCurve_P384
	// NamedCurve_P521 表示 NIST P-521。
	NamedCurve_P521
)

// ParseNamedCurve 解析不区分大小写的命名曲线名称。
func ParseNamedCurve(str string) (NamedCurve, error) {
	switch strings.ToLower(str) {
	case "none":
		return NamedCurve_None, nil
	case "x25519":
		return NamedCurve_X25519, nil
	case "p256":
		return NamedCurve_P256, nil
	case "p384":
		return NamedCurve_P384, nil
	case "p521":
		return NamedCurve_P521, nil
	default:
		return NamedCurve_None, fmt.Errorf("%w: invalid NamedCurve", ErrGTP)
	}
}

// String 返回用于配置的命名曲线名称。
func (nc NamedCurve) String() string {
	switch nc {
	case NamedCurve_X25519:
		return "x25519"
	case NamedCurve_P256:
		return "p256"
	case NamedCurve_P384:
		return "p384"
	case NamedCurve_P521:
		return "p521"
	default:
		return "none"
	}
}

// Compression 标识负载压缩算法。
type Compression uint8

const (
	// Compression_None 表示不压缩负载。
	Compression_None Compression = iota
	// Compression_Gzip 表示 Gzip。
	Compression_Gzip
	// Compression_Deflate 表示 Deflate。
	Compression_Deflate
	// Compression_Brotli 表示 Brotli。
	Compression_Brotli
	// Compression_LZ4 表示 LZ4。
	Compression_LZ4
	// Compression_Snappy 表示 Snappy。
	Compression_Snappy
)

// ParseCompression 解析不区分大小写的压缩算法名称。
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

// String 返回用于配置的压缩算法名称。
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

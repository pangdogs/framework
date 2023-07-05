package transport

import (
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
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

// AsymmetricEncryption 非对称加密算法
type AsymmetricEncryption uint8

const (
	AsymmetricEncryption_None       AsymmetricEncryption = iota // 未设置
	AsymmetricEncryption_RSA_256                                // RSA-256算法
	AsymmetricEncryption_ECDSA_P256                             // ECDSA-NIST-P256算法
)

// SymmetricEncryption 对称加密算法
type SymmetricEncryption uint8

const (
	SymmetricEncryption_None               SymmetricEncryption = iota // 未设置
	SymmetricEncryption_AES                                           // AES算法
	SymmetricEncryption_ChaCha20                                      // ChaCha20算法
	SymmetricEncryption_XChaCha20                                     // XChaCha20算法
	SymmetricEncryption_ChaCha20_Poly1305                             // ChaCha20-Poly1305算法
	SymmetricEncryption_XChaCha20_Poly1305                            // XChaCha20-Poly1305算法
)

// IV 获取iv大小和是否需要iv，iv大小为0表示与加密算法的blocksize相同
func (se SymmetricEncryption) IV() (int, bool) {
	switch se {
	case SymmetricEncryption_ChaCha20:
		return chacha20.NonceSize, true
	case SymmetricEncryption_XChaCha20:
		return chacha20.NonceSizeX, true
	default:
		return 0, false
	}
}

// Nonce 是否需要nonce，nonce大小为0表示与加密算法的blocksize相同
func (se SymmetricEncryption) Nonce() (int, bool) {
	switch se {
	case SymmetricEncryption_ChaCha20_Poly1305:
		return chacha20poly1305.NonceSize, true
	case SymmetricEncryption_XChaCha20_Poly1305:
		return chacha20poly1305.NonceSizeX, true
	default:
		return 0, false
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

// BlockCipherMode 分组密码工作模式
type BlockCipherMode uint8

const (
	BlockCipherMode_None BlockCipherMode = iota // 未设置
	BlockCipherMode_CTR                         // CTR模式
	BlockCipherMode_CBC                         // CBC模式
	BlockCipherMode_CFB                         // CFB模式
	BlockCipherMode_OFB                         // OFB模式
	BlockCipherMode_GCM                         // GCM模式
)

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
	Hash_None     Hash = iota // 未设置
	Hash_Fnv1a32              // Fnv-1a 32bit算法（用于MAC）
	Hash_Fnv1a64              // Fnv-1a 64bit算法（用于MAC）
	Hash_Fnv1a128             // Fnv-1a 128bit算法（用于MAC）
	Hash_SHA256               // SHA256算法（用于非对称加密或MAC）
)

// Bits 位数
func (h Hash) Bits() int {
	switch h {
	case Hash_Fnv1a32:
		return 32
	case Hash_Fnv1a64:
		return 64
	case Hash_Fnv1a128:
		return 128
	case Hash_SHA256:
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

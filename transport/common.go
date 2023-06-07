package transport

// Version 协议版本
type Version = uint16

const (
	Version_V1_0 Version = 0x0100 // 协议v1.0版本
)

// SecretKeyExchangeMethod 秘钥交换函数
type SecretKeyExchangeMethod = uint8

const (
	SecretKeyExchangeMethod_None  SecretKeyExchangeMethod = iota // 未设置
	SecretKeyExchangeMethod_ECDHE                                // ECDHE算法
)

// AsymmetricEncryptMethod 非对称加密函数
type AsymmetricEncryptMethod = uint8

const (
	AsymmetricEncryptMethod_None AsymmetricEncryptMethod = iota // 未设置
	AsymmetricEncryptMethod_RSA                                 // RSA算法
	AsymmetricEncryptMethod_ECC                                 // ECC算法
)

// PaddingMode 非对称加密算法填充方案
type PaddingMode = uint8

const (
	PaddingMode_None  PaddingMode = iota // 未设置
	PaddingMode_Pkcs1                    // pkcs1方案
)

// SymmetricEncryptMethod 对称加密函数
type SymmetricEncryptMethod = uint8

const (
	SymmetricEncryptMethod_None              SymmetricEncryptMethod = iota // 未设置
	SymmetricEncryptMethod_AES256                                          // AES256算法
	SymmetricEncryptMethod_ChaCha20                                        // ChaCha20算法
	SymmetricEncryptMethod_ChaCha20_Poly1305                               // ChaCha20-Poly1305算法
)

// BlockCipherMode 对称加密算法分组模式
type BlockCipherMode = uint8

const (
	BlockCipherMode_None BlockCipherMode = iota // 未设置
	BlockCipherMode_ECB                         // ECB模式
	BlockCipherMode_CBC                         // CBC模式
	BlockCipherMode_CFB                         // CFB模式
	BlockCipherMode_GCM                         // GCM模式
	BlockCipherMode_OFB                         // OFB模式
)

// HashMethod 摘要函数
type HashMethod = uint8

const (
	HashMethod_None     HashMethod = iota // 未设置
	HashMethod_Fnv1a32                    // Fnv-1a 32bit算法
	HashMethod_Fnv1a64                    // Fnv-1a 64bit算法
	HashMethod_Poly1305                   // Poly1305算法
	HashMethod_SHA256                     // SHA256算法
)

// NamedCurve 曲线类型
type NamedCurve = uint8

const (
	NamedCurve_None      = iota // 未设置
	NamedCurve_X25519           // 曲线x25519
	NamedCurve_Secp256r1        // 曲线Secp256r1
)

// CompressionMethod 压缩函数
type CompressionMethod = uint8

const (
	CompressionMethod_None   CompressionMethod = iota // 未设置
	CompressionMethod_Gzip                            // Gzip压缩算法
	CompressionMethod_Brotli                          // Brotli压缩算法
	CompressionMethod_LZ4                             // LZ4压缩算法
	CompressionMethod_Snappy                          // Snappy压缩算法
)

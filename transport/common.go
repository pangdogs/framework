package transport

// Version 协议版本
type Version = uint16

const (
	Version_V1_0 Version = 0x0100 // 协议v1.0版本
)

// SecretKeyExchange 秘钥交换函数
type SecretKeyExchange = uint8

const (
	SecretKeyExchange_None  SecretKeyExchange = iota // 未设置
	SecretKeyExchange_ECDHE                          // ECDHE算法
)

// AsymmetricEncryption 非对称加密算法
type AsymmetricEncryption = uint8

const (
	AsymmetricEncryption_None AsymmetricEncryption = iota // 未设置
	AsymmetricEncryption_RSA                              // RSA算法
	AsymmetricEncryption_ECC                              // ECC算法
)

// PaddingMode 非对称加密算法填充方案
type PaddingMode = uint8

const (
	PaddingMode_None  PaddingMode = iota // 未设置
	PaddingMode_Pkcs1                    // pkcs1方案
)

// SymmetricEncryption 对称加密算法
type SymmetricEncryption = uint8

const (
	SymmetricEncryption_None     SymmetricEncryption = iota // 未设置
	SymmetricEncryption_AES                                 // AES算法
	SymmetricEncryption_ChaCha20                            // ChaCha20算法
)

// BlockCipherMode 对称加密算法分组模式
type BlockCipherMode = uint8

const (
	BlockCipherMode_None BlockCipherMode = iota // 未设置
	BlockCipherMode_CTR                         // CTR模式
	BlockCipherMode_CBC                         // CBC模式
	BlockCipherMode_CFB                         // CFB模式
	BlockCipherMode_OFB                         // OFB模式
	BlockCipherMode_GCM                         // GCM模式
)

// Hash 摘要函数
type Hash = uint8

const (
	Hash_None     Hash = iota // 未设置
	Hash_Fnv1a32              // Fnv-1a 32bit算法
	Hash_Fnv1a64              // Fnv-1a 64bit算法
	Hash_Fnv1a128             // Fnv-1a 128bit算法
	Hash_SHA256               // SHA256算法
)

// NamedCurve 曲线类型
type NamedCurve = uint8

const (
	NamedCurve_None      = iota // 未设置
	NamedCurve_X25519           // 曲线x25519
	NamedCurve_Secp256r1        // 曲线Secp256r1
)

// Compression 压缩函数
type Compression = uint8

const (
	Compression_None    Compression = iota // 未设置
	Compression_Gzip                       // Gzip压缩算法
	Compression_Deflate                    // Deflate压缩算法
	Compression_Brotli                     // Brotli压缩算法
	Compression_LZ4                        // LZ4压缩算法
	Compression_Snappy                     // Snappy压缩算法
)

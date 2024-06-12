package codec

// CreateEncoder 创建消息包编码器
func CreateEncoder() EncoderCreator {
	return EncoderCreator{
		encoder: &Encoder{},
	}
}

// EncoderCreator 消息包编码器构建器
type EncoderCreator struct {
	encoder *Encoder
}

// SetupEncryptionModule 安装加密模块
func (ec EncoderCreator) SetupEncryptionModule(encryptionModule IEncryptionModule) EncoderCreator {
	if ec.encoder == nil {
		panic("gtp: must invoke CreateEncoder() first")
	}
	ec.encoder.EncryptionModule = encryptionModule
	ec.encoder.Encryption = encryptionModule != nil
	return ec
}

// SetupMACModule 安装MAC模块
func (ec EncoderCreator) SetupMACModule(macModule IMACModule) EncoderCreator {
	if ec.encoder == nil {
		panic("gtp: must invoke CreateEncoder() first")
	}
	ec.encoder.MACModule = macModule
	ec.encoder.PatchMAC = macModule != nil
	return ec
}

// SetupCompressionModule 安装压缩模块
func (ec EncoderCreator) SetupCompressionModule(compressionModule ICompressionModule, compressedSize int) EncoderCreator {
	if ec.encoder == nil {
		panic("gtp: must invoke CreateEncoder() first")
	}
	ec.encoder.CompressionModule = compressionModule
	if compressionModule != nil {
		ec.encoder.CompressedSize = compressedSize
	} else {
		ec.encoder.CompressedSize = 0
	}
	return ec
}

// Spawn 获取消息包编码器
func (ec EncoderCreator) Spawn() IEncoder {
	if ec.encoder == nil {
		panic("gtp: must invoke CreateEncoder() first")
	}
	return ec.encoder
}

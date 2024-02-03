package codec

import (
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
)

// CreateDecoder 创建消息包解码器
func CreateDecoder(msgCreator gtp.IMsgCreator) DecoderCreator {
	if msgCreator == nil {
		panic(fmt.Errorf("%w: msgCreator is nil", core.ErrArgs))
	}

	return DecoderCreator{
		decoder: &Decoder{
			MsgCreator: msgCreator,
		},
	}
}

// DecoderCreator 消息包解码器构建器
type DecoderCreator struct {
	decoder *Decoder
}

// SetupEncryptionModule 安装加密模块
func (dc DecoderCreator) SetupEncryptionModule(encryptionModule IEncryptionModule) DecoderCreator {
	if dc.decoder == nil {
		panic("gtp: must invoke CreateDecoder() first")
	}
	dc.decoder.EncryptionModule = encryptionModule
	return dc
}

// SetupMACModule 安装MAC模块
func (dc DecoderCreator) SetupMACModule(macModule IMACModule) DecoderCreator {
	if dc.decoder == nil {
		panic("gtp: must invoke CreateDecoder() first")
	}
	dc.decoder.MACModule = macModule
	return dc
}

// SetupCompressionModule 安装压缩模块
func (dc DecoderCreator) SetupCompressionModule(compressionModule ICompressionModule) DecoderCreator {
	if dc.decoder == nil {
		panic("gtp: must invoke CreateDecoder() first")
	}
	dc.decoder.CompressionModule = compressionModule
	return dc
}

// Spawn 获取消息包解码器
func (dc DecoderCreator) Spawn() IDecoder {
	if dc.decoder == nil {
		panic("gtp: must invoke CreateDecoder() first")
	}
	return dc.decoder
}

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

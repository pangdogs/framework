package codec

import (
	"fmt"
	"kit.golaxy.org/golaxy"
	"kit.golaxy.org/plugins/gtp"
)

// CreateDecoder 创建消息包解码器
func CreateDecoder(msgCreator gtp.IMsgCreator) _DecoderCreator {
	if msgCreator == nil {
		panic(fmt.Errorf("%w: msgCreator is nil", golaxy.ErrArgs))
	}

	return _DecoderCreator{
		decoder: &Decoder{
			MsgCreator: msgCreator,
		},
	}
}

// _DecoderCreator 消息包解码器构建器
type _DecoderCreator struct {
	decoder *Decoder
}

// SetupEncryptionModule 安装加密模块
func (dc _DecoderCreator) SetupEncryptionModule(encryptionModule IEncryptionModule) _DecoderCreator {
	dc.decoder.EncryptionModule = encryptionModule
	return dc
}

// SetupMACModule 安装MAC模块
func (dc _DecoderCreator) SetupMACModule(macModule IMACModule) _DecoderCreator {
	dc.decoder.MACModule = macModule
	return dc
}

// SetupCompressionModule 安装压缩模块
func (dc _DecoderCreator) SetupCompressionModule(compressionModule ICompressionModule) _DecoderCreator {
	dc.decoder.CompressionModule = compressionModule
	return dc
}

// Spawn 获取消息包解码器
func (dc _DecoderCreator) Spawn() IDecoder {
	return dc.decoder
}

// CreateEncoder 创建消息包编码器
func CreateEncoder() _EncoderCreator {
	return _EncoderCreator{
		encoder: &Encoder{},
	}
}

// _EncoderCreator 消息包编码器构建器
type _EncoderCreator struct {
	encoder *Encoder
}

// SetupEncryptionModule 安装加密模块
func (ec _EncoderCreator) SetupEncryptionModule(encryptionModule IEncryptionModule) _EncoderCreator {
	ec.encoder.EncryptionModule = encryptionModule
	ec.encoder.Encryption = encryptionModule != nil
	return ec
}

// SetupMACModule 安装MAC模块
func (ec _EncoderCreator) SetupMACModule(macModule IMACModule) _EncoderCreator {
	ec.encoder.MACModule = macModule
	ec.encoder.PatchMAC = macModule != nil
	return ec
}

// SetupCompressionModule 安装压缩模块
func (ec _EncoderCreator) SetupCompressionModule(compressionModule ICompressionModule, compressedSize int) _EncoderCreator {
	ec.encoder.CompressionModule = compressionModule
	if compressionModule != nil {
		ec.encoder.CompressedSize = compressedSize
	} else {
		ec.encoder.CompressedSize = 0
	}
	return ec
}

// Spawn 获取消息包编码器
func (ec _EncoderCreator) Spawn() IEncoder {
	return ec.encoder
}

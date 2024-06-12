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

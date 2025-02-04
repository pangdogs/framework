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

package codec

import (
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
)

// BuildDecoder 创建消息包解码器
func BuildDecoder(msgCreator gtp.IMsgCreator) DecoderCreator {
	if msgCreator == nil {
		exception.Panicf("%w: %w: msgCreator is nil", ErrDecode, core.ErrArgs)
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
		exception.Panicf("%w: must invoke BuildDecoder() first", ErrDecode)
	}
	dc.decoder.EncryptionModule = encryptionModule
	return dc
}

// SetupMACModule 安装MAC模块
func (dc DecoderCreator) SetupMACModule(macModule IMACModule) DecoderCreator {
	if dc.decoder == nil {
		exception.Panicf("%w: must invoke BuildDecoder() first", ErrDecode)
	}
	dc.decoder.MACModule = macModule
	return dc
}

// SetupCompressionModule 安装压缩模块
func (dc DecoderCreator) SetupCompressionModule(compressionModule ICompressionModule) DecoderCreator {
	if dc.decoder == nil {
		exception.Panicf("%w: must invoke BuildDecoder() first", ErrDecode)
	}
	dc.decoder.CompressionModule = compressionModule
	return dc
}

// Make 获取消息包解码器
func (dc DecoderCreator) Make() IDecoder {
	if dc.decoder == nil {
		exception.Panicf("%w: must invoke BuildDecoder() first", ErrDecode)
	}
	return dc.decoder
}

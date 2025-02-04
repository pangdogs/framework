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

import "git.golaxy.org/core/utils/exception"

// BuildEncoder 创建消息包编码器
func BuildEncoder() EncoderCreator {
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
		exception.Panicf("%w: must invoke BuildEncoder() first", ErrEncode)
	}
	ec.encoder.EncryptionModule = encryptionModule
	ec.encoder.Encryption = encryptionModule != nil
	return ec
}

// SetupMACModule 安装MAC模块
func (ec EncoderCreator) SetupMACModule(macModule IMACModule) EncoderCreator {
	if ec.encoder == nil {
		exception.Panicf("%w: must invoke BuildEncoder() first", ErrEncode)
	}
	ec.encoder.MACModule = macModule
	ec.encoder.PatchMAC = macModule != nil
	return ec
}

// SetupCompressionModule 安装压缩模块
func (ec EncoderCreator) SetupCompressionModule(compressionModule ICompressionModule, compressedSize int) EncoderCreator {
	if ec.encoder == nil {
		exception.Panicf("%w: must invoke BuildEncoder() first", ErrEncode)
	}
	ec.encoder.CompressionModule = compressionModule
	if compressionModule != nil {
		ec.encoder.CompressedSize = compressedSize
	} else {
		ec.encoder.CompressedSize = 0
	}
	return ec
}

// Make 获取消息包编码器
func (ec EncoderCreator) Make() IEncoder {
	if ec.encoder == nil {
		exception.Panicf("%w: must invoke BuildEncoder() first", ErrEncode)
	}
	return ec.encoder
}

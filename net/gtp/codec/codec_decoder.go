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
	"bytes"
	"errors"
	"fmt"
	"io"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
)

var (
	// ErrDecode 是 GTP 消息包解码错误的根错误。
	ErrDecode = errors.New("gtp-decode")
	// ErrUnableToPeekLength 表示输入不足以读取消息包长度。
	ErrUnableToPeekLength = fmt.Errorf("%w: %w, unable to peek length", ErrDecode, io.ErrShortBuffer)
	// ErrPacketTooLarge 表示消息包长度超过配置上限。
	ErrPacketTooLarge = fmt.Errorf("%w: packet too large", ErrDecode)
)

// NewDecoder 创建使用指定消息构建器的解码器；构建器不得为 nil。
func NewDecoder(msgCreator gtp.IMsgCreator) *Decoder {
	if msgCreator == nil {
		exception.Panicf("%w: %w: msgCreator is nil", ErrDecode, core.ErrArgs)
	}
	return &Decoder{
		MsgCreator: msgCreator,
	}
}

// IValidation 在解密和解压前校验原始消息头及消息体。
type IValidation interface {
	// Validate 校验原始消息包；返回错误会中止解码。
	Validate(msgHead gtp.MsgHead, msgBuf []byte) error
}

// Decoder 按“解密、认证、解压”的顺序还原 GTP 消息包。
type Decoder struct {
	MsgCreator          gtp.IMsgCreator // 用于构造消息体的消息构建器。
	Encryption          IEncryption     // 可选的解密模块。
	Authentication      IAuthentication // 可选的消息认证模块。
	Compression         ICompression    // 可选的解压模块。
	MaxUncompressedSize int             // 解压后负载的字节上限，用于防御压缩炸弹。
	MaxPacketSize       int             // 完整消息包的字节上限；小于等于零时不限制。
}

// SetEncryption 设置解密模块；应在开始并发解码前完成配置。
func (d *Decoder) SetEncryption(encryption IEncryption) *Decoder {
	d.Encryption = encryption
	return d
}

// SetAuthentication 设置消息认证模块；应在开始并发解码前完成配置。
func (d *Decoder) SetAuthentication(authentication IAuthentication) *Decoder {
	d.Authentication = authentication
	return d
}

// SetCompression 设置解压模块和解压后负载字节上限。
func (d *Decoder) SetCompression(compression ICompression, maxUncompressedSize int) *Decoder {
	d.Compression = compression
	d.MaxUncompressedSize = maxUncompressedSize
	return d
}

// SetMaxPacketSize 设置完整消息包字节上限；小于等于零时不限制。
func (d *Decoder) SetMaxPacketSize(maxPacketSize int) *Decoder {
	d.MaxPacketSize = maxPacketSize
	return d
}

// Decode 从 data 起始位置解码一个完整消息包，并返回消费字节数。
// 输入不足时消费字节数表示所需包长；成功返回的消息体不引用 data。
func (d *Decoder) Decode(data []byte, validation IValidation) (gtp.MsgPacket, int, error) {
	if d.MsgCreator == nil {
		return gtp.MsgPacket{}, 0, fmt.Errorf("%w: MsgCreator is nil", ErrDecode)
	}

	// 探测消息包长度
	length, err := d.peekLength(data)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	// 解码消息包
	mp, err := d.decode(data[:length], validation)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	return mp, length, nil
}

// peekLength 探测消息包长度
func (d *Decoder) peekLength(data []byte) (int, error) {
	mpl := gtp.MsgPacketLen{}

	// 读取消息包长度
	if _, err := mpl.Write(data); err != nil {
		return 0, ErrUnableToPeekLength
	}
	if d.MaxPacketSize > 0 && int(mpl.Len) > d.MaxPacketSize {
		return int(mpl.Len), fmt.Errorf("%w (%d > %d)", ErrPacketTooLarge, mpl.Len, d.MaxPacketSize)
	}
	if len(data) < int(mpl.Len) {
		return int(mpl.Len), fmt.Errorf("%w: %w (%d < %d)", ErrDecode, io.ErrShortBuffer, len(data), mpl.Len)
	}

	return int(mpl.Len), nil
}

// decode 复制并解码一个完整消息包，按标志依次执行校验、解密、认证和解压。
func (d *Decoder) decode(data []byte, validation IValidation) (gtp.MsgPacket, error) {
	// 后续变换可原地执行，因此先把调用方输入复制到池化缓冲区。
	buf := binaryutil.NewBytes(true, len(data))
	copy(buf.Payload(), data)

	mp := gtp.MsgPacket{}

	// 消息头不参与后续消息体变换。
	if _, err := mp.Head.Write(buf.Payload()); err != nil {
		buf.Release()
		return gtp.MsgPacket{}, fmt.Errorf("%w: read msg-packet-head failed, %w", ErrDecode, err)
	}

	msgBuf := buf.Payload()[mp.Head.Size():]

	// 在解密和解压前验证线路上的原始消息体。
	if validation != nil {
		if err := validation.Validate(mp.Head, msgBuf); err != nil {
			buf.Release()
			return gtp.MsgPacket{}, fmt.Errorf("%w: validate msg-packet-head failed, %w", ErrDecode, err)
		}
	}

	// 加密消息先解密，再从明文末尾校验并移除认证码。
	if mp.Head.Flags.Is(gtp.Flag_Encrypted) {
		if d.Encryption == nil {
			buf.Release()
			return gtp.MsgPacket{}, fmt.Errorf("%w: Encryption is nil, msg can't be decrypted", ErrDecode)
		}

		dencryptBuf, err := d.Encryption.Transforming(msgBuf, msgBuf)
		if err != nil {
			buf.Release()
			return gtp.MsgPacket{}, fmt.Errorf("%w: dencrypt msg failed, %w", ErrDecode, err)
		}
		if !buf.SameRef(dencryptBuf) {
			buf.Release()
		}
		buf = dencryptBuf
		msgBuf = buf.Payload()

		// 认证码覆盖解密后的消息体及消息头元数据。
		if mp.Head.Flags.Is(gtp.Flag_Signed) {
			if d.Authentication == nil {
				buf.Release()
				return gtp.MsgPacket{}, fmt.Errorf("%w: Authentication is nil, msg can't be auth msg-mac", ErrDecode)
			}
			msgBuf, err = d.Authentication.Auth(mp.Head.MsgId, mp.Head.Flags, msgBuf)
			if err != nil {
				buf.Release()
				return gtp.MsgPacket{}, fmt.Errorf("%w: auth msg-mac failed, %w", ErrDecode, err)
			}
		}
	}

	// 认证完成后再解压，并限制解压后的最大字节数。
	if mp.Head.Flags.Is(gtp.Flag_Compressed) {
		if d.Compression == nil {
			buf.Release()
			return gtp.MsgPacket{}, fmt.Errorf("%w: Compression is nil, msg can't be uncompress", ErrDecode)
		}
		uncompressedBuf, err := d.Compression.Uncompress(msgBuf, d.MaxUncompressedSize)
		if err != nil {
			buf.Release()
			return gtp.MsgPacket{}, fmt.Errorf("%w: uncompress msg failed, %w", ErrDecode, err)
		}
		if !buf.SameRef(uncompressedBuf) {
			buf.Release()
		}
		buf = uncompressedBuf
		msgBuf = buf.Payload()
	}

	// 按类型创建具体消息，未知类型在这里失败。
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		buf.Release()
		return gtp.MsgPacket{}, fmt.Errorf("%w: new msg failed, %w (%d)", ErrDecode, err, mp.Head.MsgId)
	}

	// 再复制一次，使成功返回的消息不引用即将归还池中的缓冲区。
	if _, err = msg.Write(bytes.Clone(msgBuf)); err != nil {
		buf.Release()
		return gtp.MsgPacket{}, fmt.Errorf("%w: read msg failed, %w", ErrDecode, err)
	}

	mp.Body = msg

	buf.Release()
	return mp, nil
}

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
	"errors"
	"fmt"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
)

var (
	ErrDecode               = errors.New("gtp-decode")                                                    // 解码错误
	ErrUnableToDetectLength = fmt.Errorf("%w: %w, unable to detect length", ErrDecode, io.ErrShortBuffer) // 无法探测消息长度
)

// IValidate 验证消息包接口
type IValidate interface {
	// Validate 验证消息包
	Validate(msgHead gtp.MsgHead, msgBuf []byte) error
}

// IDecoder 消息包解码器接口
type IDecoder interface {
	// Decode 解码消息包
	Decode(data []byte, validate ...IValidate) (gtp.MsgPacket, int, error)
	// GC GC
	GC()
}

// Decoder 消息包解码器
type Decoder struct {
	MsgCreator        gtp.IMsgCreator           // 消息对象构建器
	EncryptionModule  IEncryptionModule         // 加密模块
	MACModule         IMACModule                // MAC模块
	CompressionModule ICompressionModule        // 压缩模块
	gcList            []binaryutil.RecycleBytes // GC列表
}

// Decode 解码消息包
func (d *Decoder) Decode(data []byte, validate ...IValidate) (gtp.MsgPacket, int, error) {
	if d.MsgCreator == nil {
		return gtp.MsgPacket{}, 0, fmt.Errorf("%w: MsgCreator is nil", ErrDecode)
	}

	// 探测消息包长度
	length, err := d.lengthDetection(data)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	// 解码消息包
	mp, err := d.decode(data[:length], validate...)
	if err != nil {
		return gtp.MsgPacket{}, length, err
	}

	return mp, length, nil
}

// GC GC
func (d *Decoder) GC() {
	for i := range d.gcList {
		d.gcList[i].Release()
	}
	d.gcList = d.gcList[:0]
}

// lengthDetection 消息包长度探针
func (d *Decoder) lengthDetection(data []byte) (int, error) {
	mpl := gtp.MsgPacketLen{}

	// 读取消息包长度
	if _, err := mpl.Write(data); err != nil {
		return 0, ErrUnableToDetectLength
	}

	if len(data) < int(mpl.Len) {
		return int(mpl.Len), fmt.Errorf("%w: %w (%d < %d)", ErrDecode, io.ErrShortBuffer, len(data), mpl.Len)
	}

	return int(mpl.Len), nil
}

// decode 解码消息包
func (d *Decoder) decode(data []byte, validate ...IValidate) (gtp.MsgPacket, error) {
	// 消息包数据缓存
	mpBuf := binaryutil.MakeRecycleBytes(len(data))
	d.gcList = append(d.gcList, mpBuf)

	// 拷贝消息包
	copy(mpBuf.Data(), data)

	mp := gtp.MsgPacket{}

	// 读取消息头
	if _, err := mp.Head.Write(mpBuf.Data()); err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("%w: read msg-packet-head failed, %w", ErrDecode, err)
	}

	msgBuf := mpBuf.Data()[mp.Head.Size():]

	// 验证消息包
	if len(validate) > 0 {
		if err := validate[0].Validate(mp.Head, msgBuf); err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("%w: validate msg-packet-head failed, %w", ErrDecode, err)
		}
	}

	// 检查加密标记
	if mp.Head.Flags.Is(gtp.Flag_Encrypted) {
		// 解密消息体
		if d.EncryptionModule == nil {
			return gtp.MsgPacket{}, fmt.Errorf("%w: EncryptionModule is nil, msg can't be decrypted", ErrDecode)
		}
		dencryptBuf, err := d.EncryptionModule.Transforming(msgBuf, msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("%w: dencrypt msg failed, %w", ErrDecode, err)
		}
		if dencryptBuf.Recyclable() {
			d.gcList = append(d.gcList, dencryptBuf)
		}
		msgBuf = dencryptBuf.Data()

		// 检查MAC标记
		if mp.Head.Flags.Is(gtp.Flag_MAC) {
			if d.MACModule == nil {
				return gtp.MsgPacket{}, fmt.Errorf("%w: MACModule is nil, msg can't be verify MAC", ErrDecode)
			}
			// 检测MAC
			msgBuf, err = d.MACModule.VerifyMAC(mp.Head.MsgId, mp.Head.Flags, msgBuf)
			if err != nil {
				return gtp.MsgPacket{}, fmt.Errorf("%w: verify msg-mac failed, %w", ErrDecode, err)
			}
		}
	}

	// 检查压缩标记
	if mp.Head.Flags.Is(gtp.Flag_Compressed) {
		if d.CompressionModule == nil {
			return gtp.MsgPacket{}, fmt.Errorf("%w: CompressionModule is nil, msg can't be uncompress", ErrDecode)
		}
		uncompressedBuf, err := d.CompressionModule.Uncompress(msgBuf)
		if err != nil {
			return gtp.MsgPacket{}, fmt.Errorf("%w: uncompress msg failed, %w", ErrDecode, err)
		}
		if uncompressedBuf.Recyclable() {
			d.gcList = append(d.gcList, uncompressedBuf)
		}
		msgBuf = uncompressedBuf.Data()
	}

	// 创建消息体
	msg, err := d.MsgCreator.New(mp.Head.MsgId)
	if err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("%w: new msg failed, %w (%d)", ErrDecode, err, mp.Head.MsgId)
	}

	// 读取消息
	if _, err = msg.Write(msgBuf); err != nil {
		return gtp.MsgPacket{}, fmt.Errorf("%w: read msg failed, %w", ErrDecode, err)
	}

	mp.Msg = msg

	return mp, nil
}

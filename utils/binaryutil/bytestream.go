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

package binaryutil

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
)

var (
	// ErrInvalidSeekPos 表示读写游标的目标位置超出缓冲区。
	ErrInvalidSeekPos = errors.New("invalid seek position")
)

type noCopy struct{}

func (*noCopy) Lock() {}

func (*noCopy) Unlock() {}

// NewByteStream 创建直接读写 p 的字节流，读写游标均从零开始。
// endian 不得为 nil，否则 panic。
func NewByteStream(p []byte, endian binary.ByteOrder) ByteStream {
	if endian == nil {
		exception.Panicf("%w: endian is nil", core.ErrArgs)
	}
	return ByteStream{
		sp:     p,
		wp:     p,
		rp:     p,
		endian: endian,
	}
}

// NewBigEndianStream 创建使用大端字节序并直接读写 p 的字节流。
func NewBigEndianStream(p []byte) ByteStream {
	return NewByteStream(p, binary.BigEndian)
}

// NewLittleEndianStream 创建使用小端字节序并直接读写 p 的字节流。
func NewLittleEndianStream(p []byte) ByteStream {
	return NewByteStream(p, binary.LittleEndian)
}

// ByteStream 在固定字节切片上维护相互独立的读游标和写游标。
//
// ByteStream 不扩容、不复制输入，也不支持并发使用。其内部带有复制检测标记，初始化后应通过指针使用，
// 不应再按值复制。
type ByteStream struct {
	_          noCopy
	sp, wp, rp []byte
	endian     binary.ByteOrder
}

// ReadFrom 调用 reader.Read 一次，将数据读入未写区域并推进写游标；它不会循环填满缓冲区。
func (s *ByteStream) ReadFrom(reader io.Reader) (int64, error) {
	if reader == nil {
		return 0, fmt.Errorf("%w: reader is nil", core.ErrArgs)
	}
	return CopyToByteStream(s, reader)
}

// WriteTo 调用 writer.Write 一次，写出当前未读区域，并按实际写入字节数推进读游标。
func (s *ByteStream) WriteTo(writer io.Writer) (int64, error) {
	if writer == nil {
		return 0, fmt.Errorf("%w: writer is nil", core.ErrArgs)
	}
	n, err := writer.Write(s.rp)
	s.rp = s.rp[n:]
	return int64(n), err
}

// SeekWritePos 将写游标移动到相对原始缓冲区起点的 p；越界时返回 ErrInvalidSeekPos。
func (s *ByteStream) SeekWritePos(p int) error {
	if p < 0 || p > len(s.sp) {
		return ErrInvalidSeekPos
	}
	s.wp = s.sp[p:]
	return nil
}

// BytesWritten 返回写游标之前的字节数。
func (s *ByteStream) BytesWritten() int {
	return len(s.sp) - len(s.wp)
}

// BuffWritten 返回写游标之前的共享切片。
func (s *ByteStream) BuffWritten() []byte {
	return s.sp[:s.BytesWritten()]
}

// BytesUnwritten 返回从写游标到缓冲区末尾的字节数。
func (s *ByteStream) BytesUnwritten() int {
	return len(s.wp)
}

// BuffUnwritten 返回从写游标到缓冲区末尾的共享切片。
func (s *ByteStream) BuffUnwritten() []byte {
	return s.wp
}

// WriteInt8 按单字节补码写入 v 并推进写游标。
func (s *ByteStream) WriteInt8(v int8) error {
	return s.WriteUint8(uint8(v))
}

// WriteInt16 按配置字节序写入 v 并推进写游标。
func (s *ByteStream) WriteInt16(v int16) error {
	return s.WriteUint16(uint16(v))
}

// WriteInt32 按配置字节序写入 v 并推进写游标。
func (s *ByteStream) WriteInt32(v int32) error {
	return s.WriteUint32(uint32(v))
}

// WriteInt64 按配置字节序写入 v 并推进写游标。
func (s *ByteStream) WriteInt64(v int64) error {
	return s.WriteUint64(uint64(v))
}

// WriteUint8 写入 v 并推进写游标；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteUint8(v uint8) error {
	if len(s.wp) < SizeofInt8 {
		return io.ErrShortWrite
	}
	s.wp[0] = v
	s.wp = s.wp[SizeofInt8:]
	return nil
}

// WriteUint16 按配置字节序写入 v；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteUint16(v uint16) error {
	if len(s.wp) < SizeofUint16 {
		return io.ErrShortWrite
	}
	s.endian.PutUint16(s.wp, v)
	s.wp = s.wp[SizeofUint16:]
	return nil
}

// WriteUint32 按配置字节序写入 v；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteUint32(v uint32) error {
	if len(s.wp) < SizeofUint32 {
		return io.ErrShortWrite
	}
	s.endian.PutUint32(s.wp, v)
	s.wp = s.wp[SizeofUint32:]
	return nil
}

// WriteUint64 按配置字节序写入 v；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteUint64(v uint64) error {
	if len(s.wp) < SizeofUint64 {
		return io.ErrShortWrite
	}
	s.endian.PutUint64(s.wp, v)
	s.wp = s.wp[SizeofUint64:]
	return nil
}

// WriteFloat 按配置字节序写入 v 的 IEEE 754 位表示。
func (s *ByteStream) WriteFloat(v float32) error {
	if len(s.wp) < SizeofFloat {
		return io.ErrShortWrite
	}
	s.endian.PutUint32(s.wp, math.Float32bits(v))
	s.wp = s.wp[SizeofFloat:]
	return nil
}

// WriteDouble 按配置字节序写入 v 的 IEEE 754 位表示。
func (s *ByteStream) WriteDouble(v float64) error {
	if len(s.wp) < SizeofDouble {
		return io.ErrShortWrite
	}
	s.endian.PutUint64(s.wp, math.Float64bits(v))
	s.wp = s.wp[SizeofDouble:]
	return nil
}

// WriteByte 写入一个字节并推进写游标。
func (s *ByteStream) WriteByte(v byte) error {
	return s.WriteUint8(v)
}

// WriteBool 将 true 编码为 1、false 编码为 0。
func (s *ByteStream) WriteBool(v bool) error {
	if v {
		return s.WriteUint8(1)
	} else {
		return s.WriteUint8(0)
	}
}

// WriteBytes 以无符号变长整数长度前缀加原始字节的格式写入 v。
// 剩余空间不足时不写入任何数据并返回 io.ErrShortWrite。
func (s *ByteStream) WriteBytes(v []byte) error {
	if len(s.wp) < SizeofBytes(v) {
		return io.ErrShortWrite
	}
	err := s.WriteUvarint(uint64(len(v)))
	if err != nil {
		return err
	}
	if len(v) <= 0 {
		return nil
	}
	copy(s.wp, v)
	s.wp = s.wp[len(v):]
	return nil
}

// WriteString 以无符号变长整数长度前缀加 UTF-8 原始字节的格式写入 v。
// 剩余空间不足时不写入任何数据并返回 io.ErrShortWrite。
func (s *ByteStream) WriteString(v string) error {
	if len(s.wp) < SizeofString(v) {
		return io.ErrShortWrite
	}
	err := s.WriteUvarint(uint64(len(v)))
	if err != nil {
		return err
	}
	if len(v) <= 0 {
		return nil
	}
	copy(s.wp, v)
	s.wp = s.wp[len(v):]
	return nil
}

// WriteBytes16 写入 16 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes16(v []byte) error {
	if len(s.wp) < SizeofBytes16 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes16 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes16; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes16])
	}
	s.wp = s.wp[SizeofBytes16:]
	return nil
}

// WriteBytes32 写入 32 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes32(v []byte) error {
	if len(s.wp) < SizeofBytes32 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes32 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes32; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes32])
	}
	s.wp = s.wp[SizeofBytes32:]
	return nil
}

// WriteBytes64 写入 64 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes64(v []byte) error {
	if len(s.wp) < SizeofBytes64 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes64 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes64; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes64])
	}
	s.wp = s.wp[SizeofBytes64:]
	return nil
}

// WriteBytes128 写入 128 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes128(v []byte) error {
	if len(s.wp) < SizeofBytes128 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes128 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes128; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes128])
	}
	s.wp = s.wp[SizeofBytes128:]
	return nil
}

// WriteBytes160 写入 160 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes160(v []byte) error {
	if len(s.wp) < SizeofBytes160 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes160 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes160; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes160])
	}
	s.wp = s.wp[SizeofBytes160:]
	return nil
}

// WriteBytes256 写入 256 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes256(v []byte) error {
	if len(s.wp) < SizeofBytes256 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes256 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes256; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes256])
	}
	s.wp = s.wp[SizeofBytes256:]
	return nil
}

// WriteBytes512 写入 512 字节定长块；v 过短时补零，过长时截断。
func (s *ByteStream) WriteBytes512(v []byte) error {
	if len(s.wp) < SizeofBytes512 {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes512 {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes512; i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes512])
	}
	s.wp = s.wp[SizeofBytes512:]
	return nil
}

// WriteVarint 使用 binary.PutVarint 编码 v；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteVarint(v int64) error {
	if len(s.wp) < SizeofVarint(v) {
		return io.ErrShortWrite
	}
	n := binary.PutVarint(s.wp, v)
	s.wp = s.wp[n:]
	return nil
}

// WriteUvarint 使用 binary.PutUvarint 编码 v；空间不足时返回 io.ErrShortWrite。
func (s *ByteStream) WriteUvarint(v uint64) error {
	if len(s.wp) < SizeofUvarint(v) {
		return io.ErrShortWrite
	}
	n := binary.PutUvarint(s.wp, v)
	s.wp = s.wp[n:]
	return nil
}

// SeekReadPos 将读游标移动到相对原始缓冲区起点的 p；越界时返回 ErrInvalidSeekPos。
func (s *ByteStream) SeekReadPos(p int) error {
	if p < 0 || p > len(s.sp) {
		return ErrInvalidSeekPos
	}
	s.rp = s.sp[p:]
	return nil
}

// BytesRead 返回读游标之前的字节数。
func (s *ByteStream) BytesRead() int {
	return len(s.sp) - len(s.rp)
}

// BuffRead 返回读游标之前的共享切片。
func (s *ByteStream) BuffRead() []byte {
	return s.sp[:s.BytesRead()]
}

// BytesUnread 返回从读游标到缓冲区末尾的字节数。
func (s *ByteStream) BytesUnread() int {
	return len(s.rp)
}

// BuffUnread 返回从读游标到缓冲区末尾的共享切片。
func (s *ByteStream) BuffUnread() []byte {
	return s.rp
}

// ReadInt8 读取一个单字节补码整数并推进读游标。
func (s *ByteStream) ReadInt8() (int8, error) {
	v, err := s.ReadUint8()
	return int8(v), err
}

// ReadInt16 按配置字节序读取一个 int16 并推进读游标。
func (s *ByteStream) ReadInt16() (int16, error) {
	v, err := s.ReadUint16()
	return int16(v), err
}

// ReadInt32 按配置字节序读取一个 int32 并推进读游标。
func (s *ByteStream) ReadInt32() (int32, error) {
	v, err := s.ReadUint32()
	return int32(v), err
}

// ReadInt64 按配置字节序读取一个 int64 并推进读游标。
func (s *ByteStream) ReadInt64() (int64, error) {
	v, err := s.ReadUint64()
	return int64(v), err
}

// ReadUint8 读取一个 uint8；数据不足时返回 io.ErrUnexpectedEOF 且不推进读游标。
func (s *ByteStream) ReadUint8() (uint8, error) {
	if len(s.rp) < SizeofUint8 {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.rp[0]
	s.rp = s.rp[SizeofUint8:]
	return v, nil
}

// ReadUint16 按配置字节序读取一个 uint16；数据不足时不推进读游标。
func (s *ByteStream) ReadUint16() (uint16, error) {
	if len(s.rp) < SizeofUint16 {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint16(s.rp)
	s.rp = s.rp[SizeofUint16:]
	return v, nil
}

// ReadUint32 按配置字节序读取一个 uint32；数据不足时不推进读游标。
func (s *ByteStream) ReadUint32() (uint32, error) {
	if len(s.rp) < SizeofUint32 {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint32(s.rp)
	s.rp = s.rp[SizeofUint32:]
	return v, nil
}

// ReadUint64 按配置字节序读取一个 uint64；数据不足时不推进读游标。
func (s *ByteStream) ReadUint64() (uint64, error) {
	if len(s.rp) < SizeofUint64 {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint64(s.rp)
	s.rp = s.rp[SizeofUint64:]
	return v, nil
}

// ReadFloat 按配置字节序读取 IEEE 754 float32 位表示。
func (s *ByteStream) ReadFloat() (float32, error) {
	if len(s.rp) < SizeofFloat {
		return 0, io.ErrUnexpectedEOF
	}
	v := math.Float32frombits(s.endian.Uint32(s.rp))
	s.rp = s.rp[SizeofFloat:]
	return v, nil
}

// ReadDouble 按配置字节序读取 IEEE 754 float64 位表示。
func (s *ByteStream) ReadDouble() (float64, error) {
	if len(s.rp) < SizeofDouble {
		return 0, io.ErrUnexpectedEOF
	}
	v := math.Float64frombits(s.endian.Uint64(s.rp))
	s.rp = s.rp[SizeofDouble:]
	return v, nil
}

// ReadByte 读取一个字节并推进读游标。
func (s *ByteStream) ReadByte() (byte, error) {
	return s.ReadUint8()
}

// ReadBool 读取一个字节，零解码为 false，其他值解码为 true。
func (s *ByteStream) ReadBool() (bool, error) {
	b, err := s.ReadUint8()
	if err != nil {
		return false, err
	}
	if b != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// ReadBytes 读取带无符号变长整数长度前缀的字节切片，并复制载荷。
// 载荷不足时返回 io.ErrUnexpectedEOF，此时长度前缀已被消费。
func (s *ByteStream) ReadBytes() ([]byte, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if l <= 0 {
		return nil, nil
	}
	if uint64(len(s.rp)) < l {
		return nil, io.ErrUnexpectedEOF
	}
	v := make([]byte, l)
	copy(v, s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

// ReadBytesRef 读取带无符号变长整数长度前缀的字节切片，并返回底层缓冲区的共享视图。
// 返回值仅在底层缓冲区保持有效且未被修改期间可用；载荷不足时长度前缀已被消费。
func (s *ByteStream) ReadBytesRef() ([]byte, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if l <= 0 {
		return nil, nil
	}
	if uint64(len(s.rp)) < l {
		return nil, io.ErrUnexpectedEOF
	}
	v := s.rp[:l]
	s.rp = s.rp[l:]
	return v, nil
}

// ReadBytes16 读取并复制一个 16 字节定长块。
func (s *ByteStream) ReadBytes16() ([16]byte, error) {
	var v [16]byte
	if len(s.rp) < SizeofBytes16 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes16])
	s.rp = s.rp[SizeofBytes16:]
	return v, nil
}

// ReadBytes32 读取并复制一个 32 字节定长块。
func (s *ByteStream) ReadBytes32() ([32]byte, error) {
	var v [32]byte
	if len(s.rp) < SizeofBytes32 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes32])
	s.rp = s.rp[SizeofBytes32:]
	return v, nil
}

// ReadBytes64 读取并复制一个 64 字节定长块。
func (s *ByteStream) ReadBytes64() ([64]byte, error) {
	var v [64]byte
	if len(s.rp) < SizeofBytes64 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes64])
	s.rp = s.rp[SizeofBytes64:]
	return v, nil
}

// ReadBytes128 读取并复制一个 128 字节定长块。
func (s *ByteStream) ReadBytes128() ([128]byte, error) {
	var v [128]byte
	if len(s.rp) < SizeofBytes128 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes128])
	s.rp = s.rp[SizeofBytes128:]
	return v, nil
}

// ReadBytes160 读取并复制一个 160 字节定长块。
func (s *ByteStream) ReadBytes160() ([160]byte, error) {
	var v [160]byte
	if len(s.rp) < SizeofBytes160 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes160])
	s.rp = s.rp[SizeofBytes160:]
	return v, nil
}

// ReadBytes256 读取并复制一个 256 字节定长块。
func (s *ByteStream) ReadBytes256() ([256]byte, error) {
	var v [256]byte
	if len(s.rp) < SizeofBytes256 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes256])
	s.rp = s.rp[SizeofBytes256:]
	return v, nil
}

// ReadBytes512 读取并复制一个 512 字节定长块。
func (s *ByteStream) ReadBytes512() ([512]byte, error) {
	var v [512]byte
	if len(s.rp) < SizeofBytes512 {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes512])
	s.rp = s.rp[SizeofBytes512:]
	return v, nil
}

// ReadString 读取带无符号变长整数长度前缀的字符串，并复制字符串数据。
// 载荷不足时返回 io.ErrUnexpectedEOF，此时长度前缀已被消费。
func (s *ByteStream) ReadString() (string, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return "", err
	}
	if l <= 0 {
		return "", nil
	}
	if uint64(len(s.rp)) < l {
		return "", io.ErrUnexpectedEOF
	}
	v := string(s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

// ReadStringRef 读取带无符号变长整数长度前缀的字符串，且不复制底层字节。
// 返回字符串仅在底层缓冲区保持有效且未被修改期间可用；载荷不足时长度前缀已被消费。
func (s *ByteStream) ReadStringRef() (string, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return "", err
	}
	if l <= 0 {
		return "", nil
	}
	if uint64(len(s.rp)) < l {
		return "", io.ErrUnexpectedEOF
	}
	v := types.Bytes2String(s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

// ReadVarint 使用 binary.Varint 解码整数；编码不完整或溢出时返回 io.ErrUnexpectedEOF。
func (s *ByteStream) ReadVarint() (int64, error) {
	v, n := binary.Varint(s.rp)
	if n <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	s.rp = s.rp[n:]
	return v, nil
}

// ReadUvarint 使用 binary.Uvarint 解码整数；编码不完整或溢出时返回 io.ErrUnexpectedEOF。
func (s *ByteStream) ReadUvarint() (uint64, error) {
	v, n := binary.Uvarint(s.rp)
	if n <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	s.rp = s.rp[n:]
	return v, nil
}

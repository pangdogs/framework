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
	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/types"
	"io"
	"math"
)

var (
	// ErrInvalidSeekPos 调整的位置无效
	ErrInvalidSeekPos = errors.New("invalid seek position")
)

type noCopy struct{}

func (*noCopy) Lock() {}

func (*noCopy) Unlock() {}

func NewByteStream(p []byte, endian binary.ByteOrder) ByteStream {
	if endian == nil {
		panic(fmt.Errorf("%w: endian is nil", core.ErrArgs))
	}
	return ByteStream{
		sp:     p,
		wp:     p,
		rp:     p,
		endian: endian,
	}
}

func NewBigEndianStream(p []byte) ByteStream {
	return NewByteStream(p, binary.BigEndian)
}

func NewLittleEndianStream(p []byte) ByteStream {
	return NewByteStream(p, binary.LittleEndian)
}

func ReadFrom[T io.Reader](bs *ByteStream, reader T) (int64, error) {
	n, err := reader.Read(bs.wp)
	bs.wp = bs.wp[n:]
	return int64(n), err
}

type ByteStream struct {
	noCopy     noCopy
	sp, wp, rp []byte
	endian     binary.ByteOrder
}

func (s *ByteStream) ReadFrom(reader io.Reader) (int64, error) {
	if reader == nil {
		return 0, fmt.Errorf("%w: reader is nil", core.ErrArgs)
	}
	return ReadFrom(s, reader)
}

func (s *ByteStream) WriteTo(writer io.Writer) (int64, error) {
	if writer == nil {
		return 0, fmt.Errorf("%w: writer is nil", core.ErrArgs)
	}
	n, err := writer.Write(s.rp)
	s.rp = s.rp[n:]
	return int64(n), err
}

func (s *ByteStream) SeekWritePos(p int) error {
	if p < 0 || p >= len(s.sp) {
		return ErrInvalidSeekPos
	}
	s.wp = s.sp[p:]
	return nil
}

func (s *ByteStream) BytesWritten() int {
	return len(s.sp) - len(s.wp)
}

func (s *ByteStream) BuffWritten() []byte {
	return s.sp[:s.BytesWritten()]
}

func (s *ByteStream) BytesUnwritten() int {
	return len(s.wp)
}

func (s *ByteStream) BuffUnwritten() []byte {
	return s.wp
}

func (s *ByteStream) WriteInt8(v int8) error {
	return s.WriteUint8(uint8(v))
}

func (s *ByteStream) WriteInt16(v int16) error {
	return s.WriteUint16(uint16(v))
}

func (s *ByteStream) WriteInt32(v int32) error {
	return s.WriteUint32(uint32(v))
}

func (s *ByteStream) WriteInt64(v int64) error {
	return s.WriteUint64(uint64(v))
}

func (s *ByteStream) WriteUint8(v uint8) error {
	if len(s.wp) < SizeofInt8() {
		return io.ErrShortWrite
	}
	s.wp[0] = v
	s.wp = s.wp[SizeofInt8():]
	return nil
}

func (s *ByteStream) WriteUint16(v uint16) error {
	if len(s.wp) < SizeofUint16() {
		return io.ErrShortWrite
	}
	s.endian.PutUint16(s.wp, v)
	s.wp = s.wp[SizeofUint16():]
	return nil
}

func (s *ByteStream) WriteUint32(v uint32) error {
	if len(s.wp) < SizeofUint32() {
		return io.ErrShortWrite
	}
	s.endian.PutUint32(s.wp, v)
	s.wp = s.wp[SizeofUint32():]
	return nil
}

func (s *ByteStream) WriteUint64(v uint64) error {
	if len(s.wp) < SizeofUint64() {
		return io.ErrShortWrite
	}
	s.endian.PutUint64(s.wp, v)
	s.wp = s.wp[SizeofUint64():]
	return nil
}

func (s *ByteStream) WriteFloat(v float32) error {
	if len(s.wp) < SizeofFloat() {
		return io.ErrShortWrite
	}
	s.endian.PutUint32(s.wp, math.Float32bits(v))
	s.wp = s.wp[SizeofFloat():]
	return nil
}

func (s *ByteStream) WriteDouble(v float64) error {
	if len(s.wp) < SizeofDouble() {
		return io.ErrShortWrite
	}
	s.endian.PutUint64(s.wp, math.Float64bits(v))
	s.wp = s.wp[SizeofDouble():]
	return nil
}

func (s *ByteStream) WriteByte(v byte) error {
	return s.WriteUint8(v)
}

func (s *ByteStream) WriteBool(v bool) error {
	if v {
		return s.WriteUint8(1)
	} else {
		return s.WriteUint8(0)
	}
}

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

func (s *ByteStream) WriteBytes16(v []byte) error {
	if len(s.wp) < SizeofBytes16() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes16() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes16(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes16()])
	}
	s.wp = s.wp[SizeofBytes16():]
	return nil
}

func (s *ByteStream) WriteBytes32(v []byte) error {
	if len(s.wp) < SizeofBytes32() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes32() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes32(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes32()])
	}
	s.wp = s.wp[SizeofBytes32():]
	return nil
}

func (s *ByteStream) WriteBytes64(v []byte) error {
	if len(s.wp) < SizeofBytes64() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes64() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes64(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes64()])
	}
	s.wp = s.wp[SizeofBytes64():]
	return nil
}

func (s *ByteStream) WriteBytes128(v []byte) error {
	if len(s.wp) < SizeofBytes128() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes128() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes128(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes128()])
	}
	s.wp = s.wp[SizeofBytes128():]
	return nil
}

func (s *ByteStream) WriteBytes160(v []byte) error {
	if len(s.wp) < SizeofBytes160() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes160() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes160(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes160()])
	}
	s.wp = s.wp[SizeofBytes160():]
	return nil
}

func (s *ByteStream) WriteBytes256(v []byte) error {
	if len(s.wp) < SizeofBytes256() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes256() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes256(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes256()])
	}
	s.wp = s.wp[SizeofBytes256():]
	return nil
}

func (s *ByteStream) WriteBytes512(v []byte) error {
	if len(s.wp) < SizeofBytes512() {
		return io.ErrShortWrite
	}
	if len(v) < SizeofBytes512() {
		copy(s.wp, v)
		for i := len(v); i < SizeofBytes512(); i++ {
			s.wp[i] = 0
		}
	} else {
		copy(s.wp, v[:SizeofBytes512()])
	}
	s.wp = s.wp[SizeofBytes512():]
	return nil
}

func (s *ByteStream) WriteVarint(v int64) error {
	if len(s.wp) < SizeofVarint(v) {
		return io.ErrShortWrite
	}
	n := binary.PutVarint(s.wp, v)
	s.wp = s.wp[n:]
	return nil
}

func (s *ByteStream) WriteUvarint(v uint64) error {
	if len(s.wp) < SizeofUvarint(v) {
		return io.ErrShortWrite
	}
	n := binary.PutUvarint(s.wp, v)
	s.wp = s.wp[n:]
	return nil
}

func (s *ByteStream) SeekReadPos(p int) error {
	if p < 0 || p >= len(s.sp) {
		return ErrInvalidSeekPos
	}
	s.rp = s.sp[p:]
	return nil
}

func (s *ByteStream) BytesRead() int {
	return len(s.sp) - len(s.rp)
}

func (s *ByteStream) BuffRead() []byte {
	return s.sp[:s.BytesRead()]
}

func (s *ByteStream) BytesUnread() int {
	return len(s.rp)
}

func (s *ByteStream) BuffUnread() []byte {
	return s.rp
}

func (s *ByteStream) ReadInt8() (int8, error) {
	v, err := s.ReadUint8()
	return int8(v), err
}

func (s *ByteStream) ReadInt16() (int16, error) {
	v, err := s.ReadUint16()
	return int16(v), err
}

func (s *ByteStream) ReadInt32() (int32, error) {
	v, err := s.ReadUint32()
	return int32(v), err
}

func (s *ByteStream) ReadInt64() (int64, error) {
	v, err := s.ReadUint64()
	return int64(v), err
}

func (s *ByteStream) ReadUint8() (uint8, error) {
	if len(s.rp) < SizeofUint8() {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.rp[0]
	s.rp = s.rp[SizeofUint8():]
	return v, nil
}

func (s *ByteStream) ReadUint16() (uint16, error) {
	if len(s.rp) < SizeofUint16() {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint16(s.rp)
	s.rp = s.rp[SizeofUint16():]
	return v, nil
}

func (s *ByteStream) ReadUint32() (uint32, error) {
	if len(s.rp) < SizeofUint32() {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint32(s.rp)
	s.rp = s.rp[SizeofUint32():]
	return v, nil
}

func (s *ByteStream) ReadUint64() (uint64, error) {
	if len(s.rp) < SizeofUint64() {
		return 0, io.ErrUnexpectedEOF
	}
	v := s.endian.Uint64(s.rp)
	s.rp = s.rp[SizeofUint64():]
	return v, nil
}

func (s *ByteStream) ReadFloat() (float32, error) {
	if len(s.rp) < SizeofFloat() {
		return 0, io.ErrUnexpectedEOF
	}
	v := math.Float32frombits(s.endian.Uint32(s.rp))
	s.rp = s.rp[SizeofFloat():]
	return v, nil
}

func (s *ByteStream) ReadDouble() (float64, error) {
	if len(s.rp) < SizeofDouble() {
		return 0, io.ErrUnexpectedEOF
	}
	v := math.Float64frombits(s.endian.Uint64(s.rp))
	s.rp = s.rp[SizeofDouble():]
	return v, nil
}

func (s *ByteStream) ReadByte() (byte, error) {
	return s.ReadUint8()
}

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

func (s *ByteStream) ReadBytes() ([]byte, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if l <= 0 {
		return nil, nil
	}
	if len(s.rp) < int(l) {
		return nil, io.ErrUnexpectedEOF
	}
	v := make([]byte, l)
	copy(v, s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

func (s *ByteStream) ReadBytesRef() ([]byte, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if l <= 0 {
		return nil, nil
	}
	if len(s.rp) < int(l) {
		return nil, io.ErrUnexpectedEOF
	}
	v := s.rp[:l]
	s.rp = s.rp[l:]
	return v, nil
}

func (s *ByteStream) ReadBytes16() ([16]byte, error) {
	var v [16]byte
	if len(s.rp) < SizeofBytes16() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes16()])
	return v, nil
}

func (s *ByteStream) ReadBytes32() ([32]byte, error) {
	var v [32]byte
	if len(s.rp) < SizeofBytes32() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes32()])
	return v, nil
}

func (s *ByteStream) ReadBytes64() ([64]byte, error) {
	var v [64]byte
	if len(s.rp) < SizeofBytes64() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes64()])
	return v, nil
}

func (s *ByteStream) ReadBytes128() ([128]byte, error) {
	var v [128]byte
	if len(s.rp) < SizeofBytes128() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes128()])
	return v, nil
}

func (s *ByteStream) ReadBytes160() ([160]byte, error) {
	var v [160]byte
	if len(s.rp) < SizeofBytes160() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes160()])
	return v, nil
}

func (s *ByteStream) ReadBytes256() ([256]byte, error) {
	var v [256]byte
	if len(s.rp) < SizeofBytes256() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes256()])
	return v, nil
}

func (s *ByteStream) ReadBytes512() ([512]byte, error) {
	var v [512]byte
	if len(s.rp) < SizeofBytes512() {
		return v, io.ErrUnexpectedEOF
	}
	copy(v[:], s.rp[:SizeofBytes512()])
	return v, nil
}

func (s *ByteStream) ReadString() (string, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return "", err
	}
	if l <= 0 {
		return "", nil
	}
	if len(s.rp) < int(l) {
		return "", io.ErrUnexpectedEOF
	}
	v := string(s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

func (s *ByteStream) ReadStringRef() (string, error) {
	l, err := s.ReadUvarint()
	if err != nil {
		return "", err
	}
	if l <= 0 {
		return "", nil
	}
	if len(s.rp) < int(l) {
		return "", io.ErrUnexpectedEOF
	}
	v := types.Bytes2String(s.rp[:l])
	s.rp = s.rp[l:]
	return v, nil
}

func (s *ByteStream) ReadVarint() (int64, error) {
	v, n := binary.Varint(s.rp)
	if n <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	s.rp = s.rp[n:]
	return v, nil
}

func (s *ByteStream) ReadUvarint() (uint64, error) {
	v, n := binary.Uvarint(s.rp)
	if n <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	s.rp = s.rp[n:]
	return v, nil
}

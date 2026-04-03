package binaryutil

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strings"
	"testing"
)

func TestNewByteStreamPanicsWithNilEndian(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = NewByteStream(make([]byte, 1), nil)
}

func TestByteStreamReadWriteRoundTrip(t *testing.T) {
	buf := make([]byte, 128)
	stream := NewBigEndianStream(buf)

	if err := stream.WriteInt8(-3); err != nil {
		t.Fatalf("WriteInt8 failed: %v", err)
	}
	if err := stream.WriteUint16(0x1234); err != nil {
		t.Fatalf("WriteUint16 failed: %v", err)
	}
	if err := stream.WriteUint32(0x89ABCDEF); err != nil {
		t.Fatalf("WriteUint32 failed: %v", err)
	}
	if err := stream.WriteUint64(0x0102030405060708); err != nil {
		t.Fatalf("WriteUint64 failed: %v", err)
	}
	if err := stream.WriteFloat(3.25); err != nil {
		t.Fatalf("WriteFloat failed: %v", err)
	}
	if err := stream.WriteDouble(6.5); err != nil {
		t.Fatalf("WriteDouble failed: %v", err)
	}
	if err := stream.WriteBool(true); err != nil {
		t.Fatalf("WriteBool failed: %v", err)
	}
	if err := stream.WriteBool(false); err != nil {
		t.Fatalf("WriteBool failed: %v", err)
	}
	if err := stream.WriteBytes([]byte("abc")); err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}
	if err := stream.WriteString("xyz"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if err := stream.WriteVarint(-99); err != nil {
		t.Fatalf("WriteVarint failed: %v", err)
	}
	if err := stream.WriteUvarint(123); err != nil {
		t.Fatalf("WriteUvarint failed: %v", err)
	}

	if err := stream.SeekReadPos(0); err != nil {
		t.Fatalf("SeekReadPos failed: %v", err)
	}

	if got, err := stream.ReadInt8(); err != nil || got != -3 {
		t.Fatalf("ReadInt8 got %d err %v", got, err)
	}
	if got, err := stream.ReadUint16(); err != nil || got != 0x1234 {
		t.Fatalf("ReadUint16 got %d err %v", got, err)
	}
	if got, err := stream.ReadUint32(); err != nil || got != 0x89ABCDEF {
		t.Fatalf("ReadUint32 got %d err %v", got, err)
	}
	if got, err := stream.ReadUint64(); err != nil || got != 0x0102030405060708 {
		t.Fatalf("ReadUint64 got %d err %v", got, err)
	}
	if got, err := stream.ReadFloat(); err != nil || math.Abs(float64(got-3.25)) > 1e-6 {
		t.Fatalf("ReadFloat got %f err %v", got, err)
	}
	if got, err := stream.ReadDouble(); err != nil || math.Abs(got-6.5) > 1e-9 {
		t.Fatalf("ReadDouble got %f err %v", got, err)
	}
	if got, err := stream.ReadBool(); err != nil || !got {
		t.Fatalf("ReadBool(true) got %v err %v", got, err)
	}
	if got, err := stream.ReadBool(); err != nil || got {
		t.Fatalf("ReadBool(false) got %v err %v", got, err)
	}
	if got, err := stream.ReadBytes(); err != nil || string(got) != "abc" {
		t.Fatalf("ReadBytes got %q err %v", got, err)
	}
	if got, err := stream.ReadString(); err != nil || got != "xyz" {
		t.Fatalf("ReadString got %q err %v", got, err)
	}
	if got, err := stream.ReadVarint(); err != nil || got != -99 {
		t.Fatalf("ReadVarint got %d err %v", got, err)
	}
	if got, err := stream.ReadUvarint(); err != nil || got != 123 {
		t.Fatalf("ReadUvarint got %d err %v", got, err)
	}
}

func TestByteStreamEndian(t *testing.T) {
	buf := make([]byte, 2)
	stream := NewLittleEndianStream(buf)

	if err := stream.WriteUint16(0x1234); err != nil {
		t.Fatalf("WriteUint16 failed: %v", err)
	}
	if want := []byte{0x34, 0x12}; !bytes.Equal(buf, want) {
		t.Fatalf("unexpected bytes: got %v want %v", buf, want)
	}

	stream = NewByteStream(buf, binary.LittleEndian)
	if got, err := stream.ReadUint16(); err != nil || got != 0x1234 {
		t.Fatalf("ReadUint16 got %x err %v", got, err)
	}
}

func TestByteStreamSeekAndBuffers(t *testing.T) {
	stream := NewBigEndianStream(make([]byte, 6))

	if err := stream.WriteBytes16([]byte("1234567890123456")); err == nil {
		t.Fatal("expected short write")
	}
	if err := stream.WriteUint16(0xABCD); err != nil {
		t.Fatalf("WriteUint16 failed: %v", err)
	}
	if got := stream.BytesWritten(); got != 2 {
		t.Fatalf("unexpected bytes written: got %d want 2", got)
	}
	if want := []byte{0xAB, 0xCD}; !bytes.Equal(stream.BuffWritten(), want) {
		t.Fatalf("unexpected written buffer: got %v want %v", stream.BuffWritten(), want)
	}

	if err := stream.SeekWritePos(len(stream.sp)); err != nil {
		t.Fatalf("SeekWritePos(end) failed: %v", err)
	}
	if got := stream.BytesUnwritten(); got != 0 {
		t.Fatalf("unexpected unwritten bytes: got %d want 0", got)
	}
	if err := stream.SeekWritePos(-1); err != ErrInvalidSeekPos {
		t.Fatalf("unexpected seek write error: %v", err)
	}
	if err := stream.SeekReadPos(len(stream.sp) + 1); err != ErrInvalidSeekPos {
		t.Fatalf("unexpected seek read error: %v", err)
	}
}

func TestByteStreamReadFromAndWriteTo(t *testing.T) {
	stream := NewBigEndianStream(make([]byte, 8))

	if _, err := stream.ReadFrom(nil); err == nil {
		t.Fatal("expected nil reader error")
	}
	if _, err := stream.WriteTo(nil); err == nil {
		t.Fatal("expected nil writer error")
	}

	n, err := stream.ReadFrom(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected ReadFrom bytes: got %d want 5", n)
	}

	if err := stream.SeekReadPos(0); err != nil {
		t.Fatalf("SeekReadPos failed: %v", err)
	}

	var out bytes.Buffer
	n, err = stream.WriteTo(&out)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != int64(len(stream.sp)) {
		t.Fatalf("unexpected WriteTo bytes: got %d want %d", n, len(stream.sp))
	}
	if got := out.Bytes(); !bytes.Equal(got[:5], []byte("hello")) {
		t.Fatalf("unexpected output prefix: got %q", got[:5])
	}
}

func TestByteStreamBytesAndStringRefs(t *testing.T) {
	stream := NewBigEndianStream(make([]byte, 32))

	if err := stream.WriteBytes([]byte("abc")); err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}
	if err := stream.WriteString("xyz"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	if err := stream.SeekReadPos(0); err != nil {
		t.Fatalf("SeekReadPos failed: %v", err)
	}

	refBytes, err := stream.ReadBytesRef()
	if err != nil {
		t.Fatalf("ReadBytesRef failed: %v", err)
	}
	refString, err := stream.ReadStringRef()
	if err != nil {
		t.Fatalf("ReadStringRef failed: %v", err)
	}

	refBytes[0] = 'A'
	if refString != "xyz" {
		t.Fatalf("unexpected string ref: got %q want %q", refString, "xyz")
	}
	if got := stream.sp[1]; got != 'A' {
		t.Fatalf("expected referenced bytes to share storage, got %q", got)
	}
}

func TestByteStreamFixedBytesAdvance(t *testing.T) {
	cases := []struct {
		name string
		size int
		read func(*ByteStream) ([]byte, error)
	}{
		{
			name: "16",
			size: SizeofBytes16,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes16()
				return v[:], err
			},
		},
		{
			name: "32",
			size: SizeofBytes32,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes32()
				return v[:], err
			},
		},
		{
			name: "64",
			size: SizeofBytes64,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes64()
				return v[:], err
			},
		},
		{
			name: "128",
			size: SizeofBytes128,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes128()
				return v[:], err
			},
		},
		{
			name: "160",
			size: SizeofBytes160,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes160()
				return v[:], err
			},
		},
		{
			name: "256",
			size: SizeofBytes256,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes256()
				return v[:], err
			},
		},
		{
			name: "512",
			size: SizeofBytes512,
			read: func(s *ByteStream) ([]byte, error) {
				v, err := s.ReadBytes512()
				return v[:], err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := append(bytes.Repeat([]byte{'a'}, tc.size), bytes.Repeat([]byte{'b'}, tc.size)...)
			stream := NewBigEndianStream(buf)

			first, err := tc.read(&stream)
			if err != nil {
				t.Fatalf("first read failed: %v", err)
			}
			second, err := tc.read(&stream)
			if err != nil {
				t.Fatalf("second read failed: %v", err)
			}

			if !bytes.Equal(first, bytes.Repeat([]byte{'a'}, tc.size)) {
				t.Fatalf("unexpected first read")
			}
			if !bytes.Equal(second, bytes.Repeat([]byte{'b'}, tc.size)) {
				t.Fatalf("unexpected second read")
			}
			if got := stream.BytesRead(); got != tc.size*2 {
				t.Fatalf("unexpected bytes read: got %d want %d", got, tc.size*2)
			}
		})
	}
}

func TestByteStreamReadErrors(t *testing.T) {
	stream := NewBigEndianStream([]byte{})

	if _, err := stream.ReadUint8(); err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected ReadUint8 error: %v", err)
	}
	if _, err := stream.ReadVarint(); err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected ReadVarint error: %v", err)
	}
	if _, err := stream.ReadUvarint(); err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected ReadUvarint error: %v", err)
	}
}

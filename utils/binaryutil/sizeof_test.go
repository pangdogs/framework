package binaryutil

import "testing"

func TestSizeConstants(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{name: "int8", got: SizeofInt8, want: 1},
		{name: "int16", got: SizeofInt16, want: 2},
		{name: "int32", got: SizeofInt32, want: 4},
		{name: "int64", got: SizeofInt64, want: 8},
		{name: "uint8", got: SizeofUint8, want: 1},
		{name: "uint16", got: SizeofUint16, want: 2},
		{name: "uint32", got: SizeofUint32, want: 4},
		{name: "uint64", got: SizeofUint64, want: 8},
		{name: "float", got: SizeofFloat, want: 4},
		{name: "double", got: SizeofDouble, want: 8},
		{name: "byte", got: SizeofByte, want: 1},
		{name: "bool", got: SizeofBool, want: 1},
		{name: "bytes16", got: SizeofBytes16, want: 16},
		{name: "bytes32", got: SizeofBytes32, want: 32},
		{name: "bytes64", got: SizeofBytes64, want: 64},
		{name: "bytes128", got: SizeofBytes128, want: 128},
		{name: "bytes160", got: SizeofBytes160, want: 160},
		{name: "bytes256", got: SizeofBytes256, want: 256},
		{name: "bytes512", got: SizeofBytes512, want: 512},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("unexpected size: got %d want %d", tc.got, tc.want)
			}
		})
	}
}

func TestSizeofBytesAndString(t *testing.T) {
	cases := []struct {
		name string
		len  int
		want int
	}{
		{name: "empty", len: 0, want: 1},
		{name: "one_byte", len: 1, want: 2},
		{name: "uvarint_boundary_127", len: 127, want: 128},
		{name: "uvarint_boundary_128", len: 128, want: 130},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]byte, tc.len)
			str := string(make([]byte, tc.len))

			if got := SizeofBytes(buf); got != tc.want {
				t.Fatalf("SizeofBytes got %d want %d", got, tc.want)
			}
			if got := SizeofString(str); got != tc.want {
				t.Fatalf("SizeofString got %d want %d", got, tc.want)
			}
		})
	}
}

func TestSizeofUvarint(t *testing.T) {
	cases := []struct {
		value uint64
		want  int
	}{
		{value: 0, want: 1},
		{value: 1, want: 1},
		{value: 127, want: 1},
		{value: 128, want: 2},
		{value: 16383, want: 2},
		{value: 16384, want: 3},
		{value: 1<<21 - 1, want: 3},
		{value: 1 << 21, want: 4},
		{value: 1<<28 - 1, want: 4},
		{value: 1 << 28, want: 5},
		{value: 1<<35 - 1, want: 5},
		{value: 1 << 35, want: 6},
		{value: 1<<42 - 1, want: 6},
		{value: 1 << 42, want: 7},
		{value: 1<<49 - 1, want: 7},
		{value: 1 << 49, want: 8},
		{value: 1<<56 - 1, want: 8},
		{value: 1 << 56, want: 9},
		{value: ^uint64(0), want: 10},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			if got := SizeofUvarint(tc.value); got != tc.want {
				t.Fatalf("SizeofUvarint(%d) got %d want %d", tc.value, got, tc.want)
			}
		})
	}
}

func TestSizeofVarint(t *testing.T) {
	cases := []struct {
		value int64
		want  int
	}{
		{value: 0, want: 1},
		{value: 1, want: 1},
		{value: -1, want: 1},
		{value: 63, want: 1},
		{value: -64, want: 1},
		{value: 64, want: 2},
		{value: -65, want: 2},
		{value: 8191, want: 2},
		{value: -8192, want: 2},
		{value: 8192, want: 3},
		{value: -8193, want: 3},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			if got := SizeofVarint(tc.value); got != tc.want {
				t.Fatalf("SizeofVarint(%d) got %d want %d", tc.value, got, tc.want)
			}
		})
	}
}

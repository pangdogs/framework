package binaryutil

func SizeofInt8() int {
	return 1
}

func SizeofInt16() int {
	return 2
}

func SizeofInt32() int {
	return 4
}

func SizeofInt64() int {
	return 8
}

func SizeofUint8() int {
	return 1
}

func SizeofUint16() int {
	return 2
}

func SizeofUint32() int {
	return 4
}

func SizeofUint64() int {
	return 8
}

func SizeofByte() int {
	return 1
}

func SizeofBool() int {
	return 1
}

func SizeofBytes(v []byte) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

func SizeofString(v string) int {
	l := uint64(len(v))
	return SizeofUvarint(l) + len(v)
}

func SizeofBytes16() int {
	return 16
}

func SizeofBytes32() int {
	return 32
}

func SizeofBytes64() int {
	return 64
}

func SizeofBytes128() int {
	return 128
}

func SizeofBytes160() int {
	return 160
}

func SizeofBytes256() int {
	return 256
}

func SizeofBytes512() int {
	return 512
}

func SizeofVarint(v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}
	return SizeofUvarint(uv)
}

func SizeofUvarint(v uint64) int {
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}

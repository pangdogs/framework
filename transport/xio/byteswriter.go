package xio

// BytesWriter will only write bytes to the underlying writer until the limit is reached.
type BytesWriter struct {
	N     int
	Bytes []byte
}

func NewBytesWriter(bs []byte) *BytesWriter {
	return &BytesWriter{
		N:     0,
		Bytes: bs,
	}
}

func (l *BytesWriter) Write(p []byte) (int, error) {
	if l.N >= len(l.Bytes) {
		return 0, ErrLimitReached
	}

	// Write 0 bytes if the limit is to be exceeded.
	if len(p) > len(l.Bytes)-l.N {
		return 0, ErrLimitReached
	}

	copy(l.Bytes[l.N:], p)
	l.N += len(p)

	return len(p), nil
}

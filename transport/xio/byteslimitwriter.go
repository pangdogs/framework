package xio

// BytesLimitWriter will only write bytes to the underlying writer until the limit is reached.
type BytesLimitWriter struct {
	Limit int
	N     int
	Bs    []byte
}

func NewBytesLimitWriter(bs []byte, n int) *BytesLimitWriter {
	// If anyone tries this, just make a 0 writer.
	if n < 0 {
		n = 0
	}
	if n > len(bs) {
		n = len(bs)
	}
	return &BytesLimitWriter{
		Limit: n,
		N:     0,
		Bs:    bs,
	}
}

func (l *BytesLimitWriter) Write(p []byte) (int, error) {
	if l.N >= l.Limit {
		return 0, ErrLimitReached
	}

	// Write 0 bytes if the limit is to be exceeded.
	if len(p) > l.Limit-l.N {
		return 0, ErrLimitReached
	}

	copy(l.Bs[l.N:], p)
	l.N += len(p)

	return len(p), nil
}

package method

import (
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"io"
	"kit.golaxy.org/plugins/transport"
)

var (
	GzipCompressionLevel   = gzip.DefaultCompression
	BrotliCompressionLevel = brotli.DefaultCompression
	LZ4CompressionLevel    = lz4.Level4
)

type CompressionStream interface {
	WrapReader(r io.Reader) (io.Reader, error)
	WrapWriter(w io.Writer) (io.WriteCloser, error)
}

func NewCompressionStream(m transport.CompressionMethod) (CompressionStream, error) {
	switch m {
	case transport.CompressionMethod_Gzip:
		return &_GzipStream{}, nil
	case transport.CompressionMethod_Brotli:
		return &_BrotliStream{}, nil
	case transport.CompressionMethod_LZ4:
		return &_LZ4Stream{}, nil
	case transport.CompressionMethod_Snappy:
		return &_SnappyStream{}, nil
	default:
		return nil, ErrInvalidMethod
	}
}

type _GzipStream struct {
	reader *gzip.Reader
	writer *gzip.Writer
}

func (s *_GzipStream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.reader == nil {
		cr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		s.reader = cr
	} else {
		err := s.reader.Reset(r)
		if err != nil {
			return nil, err
		}
	}
	return s.reader, nil
}

func (s *_GzipStream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.writer == nil {
		cw, err := gzip.NewWriterLevel(w, GzipCompressionLevel)
		if err != nil {
			return nil, err
		}
		s.writer = cw
	} else {
		s.writer.Reset(w)
	}
	return s.writer, nil
}

type _BrotliStream struct {
	reader *brotli.Reader
	writer *brotli.Writer
}

func (s *_BrotliStream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.reader == nil {
		s.reader = brotli.NewReader(r)
	} else {
		err := s.reader.Reset(r)
		if err != nil {
			return nil, err
		}
	}
	return s.reader, nil
}

func (s *_BrotliStream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.writer == nil {
		s.writer = brotli.NewWriterLevel(w, BrotliCompressionLevel)
	} else {
		s.writer.Reset(w)
	}
	return s.writer, nil
}

type _LZ4Stream struct {
	reader *lz4.Reader
	writer *lz4.Writer
}

func (s *_LZ4Stream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.reader == nil {
		s.reader = lz4.NewReader(r)
	} else {
		s.reader.Reset(r)
	}
	return s.reader, nil
}

func (s *_LZ4Stream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.writer == nil {
		s.writer = lz4.NewWriter(w)
		s.writer.Apply(lz4.CompressionLevelOption(LZ4CompressionLevel))
	} else {
		s.writer.Reset(w)
	}
	return s.writer, nil
}

type _SnappyStream struct {
	reader *snappy.Reader
	writer *snappy.Writer
}

func (s *_SnappyStream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.reader == nil {
		s.reader = snappy.NewReader(r)
	} else {
		s.reader.Reset(r)
	}
	return s.reader, nil
}

func (s *_SnappyStream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.writer == nil {
		s.writer = snappy.NewWriter(w)
	} else {
		s.writer.Reset(w)
	}
	return s.writer, nil
}

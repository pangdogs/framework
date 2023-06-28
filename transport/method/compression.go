package method

import (
	"compress/flate"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"io"
	"kit.golaxy.org/plugins/transport"
)

var (
	GzipCompressionLevel    = gzip.DefaultCompression   // gzip压缩级别
	DeflateCompressionLevel = flate.DefaultCompression  // deflate压缩级别
	BrotliCompressionLevel  = brotli.DefaultCompression // brotli压缩级别
	LZ4CompressionLevel     = lz4.Level4                // lz4压缩级别
)

// CompressionStream 压缩/解压缩流
type CompressionStream interface {
	WrapReader(r io.Reader) (io.Reader, error)
	WrapWriter(w io.Writer) (io.WriteCloser, error)
}

// NewCompressionStream 创建压缩/解压缩流
func NewCompressionStream(c transport.Compression) (CompressionStream, error) {
	switch c {
	case transport.Compression_Gzip:
		return &_GzipStream{}, nil
	case transport.Compression_Deflate:
		return &_DeflateStream{}, nil
	case transport.Compression_Brotli:
		return &_BrotliStream{}, nil
	case transport.Compression_LZ4:
		return &_LZ4Stream{}, nil
	case transport.Compression_Snappy:
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

type _DeflateStream struct {
	reader   io.Reader
	resetter flate.Resetter
	writer   *flate.Writer
}

func (s *_DeflateStream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.reader == nil {
		s.reader = flate.NewReader(r)
		s.resetter = s.reader.(flate.Resetter)
	} else {
		s.resetter.Reset(r, nil)
	}
	return s.reader, nil
}

func (s *_DeflateStream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.writer == nil {
		cw, err := flate.NewWriter(w, DeflateCompressionLevel)
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

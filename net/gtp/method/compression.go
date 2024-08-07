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

package method

import (
	"compress/flate"
	"compress/gzip"
	"git.golaxy.org/framework/net/gtp"
	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"io"
)

var (
	GzipCompressionLevel    = gzip.DefaultCompression   // gzip压缩级别
	DeflateCompressionLevel = flate.DefaultCompression  // deflate压缩级别
	BrotliCompressionLevel  = brotli.DefaultCompression // brotli压缩级别
	LZ4CompressionLevel     = lz4.Level4                // lz4压缩级别
)

// CompressionStream 压缩/解压缩流
type CompressionStream interface {
	// WrapReader 包装解压缩流
	WrapReader(r io.Reader) (io.Reader, error)
	// WrapWriter 包装压缩流
	WrapWriter(w io.Writer) (io.WriteCloser, error)
}

// NewCompressionStream 创建压缩/解压缩流
func NewCompressionStream(c gtp.Compression) (CompressionStream, error) {
	switch c {
	case gtp.Compression_Gzip:
		return &_GzipStream{}, nil
	case gtp.Compression_Deflate:
		return &_DeflateStream{}, nil
	case gtp.Compression_Brotli:
		return &_BrotliStream{}, nil
	case gtp.Compression_LZ4:
		return &_LZ4Stream{}, nil
	case gtp.Compression_Snappy:
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

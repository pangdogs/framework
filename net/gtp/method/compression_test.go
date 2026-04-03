package method

import (
	"bytes"
	"io"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewCompressionStream(t *testing.T) {
	cases := []gtp.Compression{
		gtp.Compression_Gzip,
		gtp.Compression_Deflate,
		gtp.Compression_Brotli,
		gtp.Compression_LZ4,
		gtp.Compression_Snappy,
	}

	for _, tc := range cases {
		t.Run(tc.String(), func(t *testing.T) {
			stream, err := NewCompressionStream(tc)
			if err != nil {
				t.Fatalf("NewCompressionStream failed: %v", err)
			}

			src := bytes.Repeat([]byte("compression-"), 64)
			var compressed bytes.Buffer

			w, err := stream.WrapWriter(&compressed)
			if err != nil {
				t.Fatalf("WrapWriter failed: %v", err)
			}
			if _, err := w.Write(src); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Close failed: %v", err)
			}

			r, err := stream.WrapReader(bytes.NewReader(compressed.Bytes()))
			if err != nil {
				t.Fatalf("WrapReader failed: %v", err)
			}
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}

			if !bytes.Equal(got, src) {
				t.Fatal("unexpected round-trip payload")
			}
		})
	}
}

func TestNewCompressionStreamInvalid(t *testing.T) {
	if _, err := NewCompressionStream(gtp.Compression(255)); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
}

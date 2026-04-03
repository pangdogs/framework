package codec

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewCompressionPanicsWithNilStream(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	NewCompression(nil)
}

func TestCompressionCompressAndUncompressRoundTrip(t *testing.T) {
	c := newTestCompression(t)
	src := bytes.Repeat([]byte("compress-me-"), 64)

	compressed, ok, err := c.Compress(src)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	defer compressed.Release()
	if !ok {
		t.Fatal("expected data to be compressed")
	}

	uncompressed, err := c.Uncompress(compressed.Payload(), len(src)+1)
	if err != nil {
		t.Fatalf("Uncompress failed: %v", err)
	}
	defer uncompressed.Release()

	if !bytes.Equal(uncompressed.Payload(), src) {
		t.Fatal("unexpected uncompressed payload")
	}
}

func TestCompressionCompressNoBenefit(t *testing.T) {
	c := newTestCompression(t)

	compressed, ok, err := c.Compress([]byte("abc"))
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	defer compressed.Release()
	if ok {
		t.Fatal("expected compression to be skipped")
	}
}

func TestCompressionErrors(t *testing.T) {
	t.Run("nil stream receiver", func(t *testing.T) {
		c := &Compression{}
		if _, _, err := c.Compress([]byte("x")); err == nil {
			t.Fatal("expected Compress error")
		}
		if _, err := c.Uncompress([]byte("x"), 1); err == nil {
			t.Fatal("expected Uncompress error")
		}
	})

	t.Run("empty src", func(t *testing.T) {
		c := newTestCompression(t)
		if _, err := c.Uncompress(nil, 1); err == nil {
			t.Fatal("expected Uncompress error")
		}
	})

	t.Run("negative original size", func(t *testing.T) {
		c := newTestCompression(t)
		data := mustMarshalCompressed(t, gtp.MsgCompressed{Data: []byte("x"), OriginalSize: -1})
		if _, err := c.Uncompress(data, 1024); err == nil {
			t.Fatal("expected negative size error")
		}
	})

	t.Run("too large", func(t *testing.T) {
		c := newTestCompression(t)
		data := mustMarshalCompressed(t, gtp.MsgCompressed{Data: []byte("x"), OriginalSize: 10})
		if _, err := c.Uncompress(data, 5); err == nil {
			t.Fatal("expected size too large error")
		}
	})

	t.Run("wrap writer error", func(t *testing.T) {
		c := &Compression{CompressionStream: stubCompressionStream{
			wrapWriter: func(io.Writer) (io.WriteCloser, error) { return nil, errTest },
		}}
		if _, _, err := c.Compress([]byte("hello")); !errors.Is(err, errTest) {
			t.Fatalf("expected wrapped writer error, got %v", err)
		}
	})

	t.Run("wrap reader error", func(t *testing.T) {
		c := &Compression{CompressionStream: stubCompressionStream{
			wrapReader: func(io.Reader) (io.Reader, error) { return nil, errTest },
		}}
		data := mustMarshalCompressed(t, gtp.MsgCompressed{Data: []byte("x"), OriginalSize: 1})
		if _, err := c.Uncompress(data, 10); !errors.Is(err, errTest) {
			t.Fatalf("expected wrapped reader error, got %v", err)
		}
	})
}

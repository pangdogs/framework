package codec

import (
	"bytes"
	"errors"
	"hash"
	"io"
	"testing"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/utils/binaryutil"
)

func newTestPayload() *gtp.MsgPayload {
	return &gtp.MsgPayload{Data: bytes.Repeat([]byte("payload-"), 64)}
}

func newTestHMAC(tb testing.TB) hash.Hash {
	tb.Helper()

	h, err := method.NewHMAC(gtp.Hash_SHA256, bytes.Repeat([]byte{1}, 16))
	if err != nil {
		tb.Fatalf("NewHMAC failed: %v", err)
	}
	return h
}

func newTestCompression(tb testing.TB) ICompression {
	tb.Helper()

	cs, err := method.NewCompressionStream(gtp.Compression_Gzip)
	if err != nil {
		tb.Fatalf("NewCompressionStream failed: %v", err)
	}
	return NewCompression(cs)
}

func newTestEncryptionPair(tb testing.TB) (IEncryption, IEncryption) {
	tb.Helper()

	key := bytes.Repeat([]byte{2}, 16)
	nonce := bytes.Repeat([]byte{3}, 16)

	encrypter, decrypter, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key, nil, nil)
	if err != nil {
		tb.Fatalf("NewCipher failed: %v", err)
	}

	fetchNonce := func() ([]byte, error) {
		return bytes.Clone(nonce), nil
	}

	return NewEncryption(encrypter, nil, fetchNonce), NewEncryption(decrypter, nil, fetchNonce)
}

func mustEncode(tb testing.TB, encoder *Encoder, msg gtp.ReadableMsg) binaryutil.Bytes {
	tb.Helper()

	buf, err := encoder.Encode(gtp.Flags_None(), msg)
	if err != nil {
		tb.Fatalf("Encode failed: %v", err)
	}
	return buf
}

func mustMarshalCompressed(tb testing.TB, m gtp.MsgCompressed) []byte {
	tb.Helper()

	buf := make([]byte, m.Size())
	if _, err := binaryutil.CopyToBuff(buf, m); err != nil {
		tb.Fatalf("marshal compressed msg failed: %v", err)
	}
	return buf
}

type stubValidation struct {
	err error
}

func (v stubValidation) Validate(gtp.MsgHead, []byte) error {
	return v.err
}

type stubMsgCreator struct {
	newFn func(gtp.MsgId) (gtp.Msg, error)
}

func (stubMsgCreator) Declare(gtp.Msg) {}

func (c stubMsgCreator) New(msgId gtp.MsgId) (gtp.Msg, error) {
	return c.newFn(msgId)
}

type failingMsg struct {
	writeErr error
}

func (f *failingMsg) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (f *failingMsg) Write([]byte) (int, error) {
	return 0, f.writeErr
}

func (*failingMsg) Size() int {
	return 0
}

func (*failingMsg) MsgId() gtp.MsgId {
	return gtp.MsgId_Payload
}

func (f *failingMsg) Clone() gtp.Msg {
	return &failingMsg{writeErr: f.writeErr}
}

type stubCompression struct {
	compressFn   func([]byte) (binaryutil.Bytes, bool, error)
	uncompressFn func([]byte, int) (binaryutil.Bytes, error)
}

func (s stubCompression) Compress(src []byte) (binaryutil.Bytes, bool, error) {
	if s.compressFn != nil {
		return s.compressFn(src)
	}
	return binaryutil.EmptyBytes, false, nil
}

func (s stubCompression) Uncompress(src []byte, max int) (binaryutil.Bytes, error) {
	if s.uncompressFn != nil {
		return s.uncompressFn(src, max)
	}
	return binaryutil.RefBytes(src), nil
}

type stubAuthentication struct {
	signFn           func(gtp.MsgId, gtp.Flags, []byte) (binaryutil.Bytes, error)
	authFn           func(gtp.MsgId, gtp.Flags, []byte) ([]byte, error)
	sizeOfAdditionFn func(int) (int, error)
}

func (s stubAuthentication) Sign(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (binaryutil.Bytes, error) {
	if s.signFn != nil {
		return s.signFn(msgId, flags, msgBuf)
	}
	return binaryutil.RefBytes(msgBuf), nil
}

func (s stubAuthentication) Auth(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) ([]byte, error) {
	if s.authFn != nil {
		return s.authFn(msgId, flags, msgBuf)
	}
	return msgBuf, nil
}

func (s stubAuthentication) SizeOfAddition(msgLen int) (int, error) {
	if s.sizeOfAdditionFn != nil {
		return s.sizeOfAdditionFn(msgLen)
	}
	return 0, nil
}

type stubEncryption struct {
	transformFn      func([]byte, []byte) (binaryutil.Bytes, error)
	sizeOfAdditionFn func(int) (int, error)
}

func (s stubEncryption) Transforming(dst, src []byte) (binaryutil.Bytes, error) {
	if s.transformFn != nil {
		return s.transformFn(dst, src)
	}
	copy(dst, src)
	return binaryutil.RefBytes(dst[:len(src)]), nil
}

func (s stubEncryption) SizeOfAddition(msgLen int) (int, error) {
	if s.sizeOfAdditionFn != nil {
		return s.sizeOfAdditionFn(msgLen)
	}
	return 0, nil
}

type stubCompressionStream struct {
	wrapReader func(io.Reader) (io.Reader, error)
	wrapWriter func(io.Writer) (io.WriteCloser, error)
}

func (s stubCompressionStream) WrapReader(r io.Reader) (io.Reader, error) {
	if s.wrapReader != nil {
		return s.wrapReader(r)
	}
	return r, nil
}

func (s stubCompressionStream) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	if s.wrapWriter != nil {
		return s.wrapWriter(w)
	}
	return nopWriteCloser{Writer: w}, nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

type stubCipher struct {
	blockSize  int
	nonceSize  int
	overhead   int
	pad        bool
	unpad      bool
	inputSize  func(int) int
	outputSize func(int) int
	transform  func([]byte, []byte, []byte) (int, error)
}

func (c stubCipher) Transforming(dst, src, nonce []byte) (int, error) {
	if c.transform != nil {
		return c.transform(dst, src, nonce)
	}
	copy(dst, src)
	return len(src), nil
}

func (c stubCipher) BlockSize() int {
	return c.blockSize
}

func (c stubCipher) NonceSize() int {
	return c.nonceSize
}

func (c stubCipher) Overhead() int {
	return c.overhead
}

func (c stubCipher) Pad() bool {
	return c.pad
}

func (c stubCipher) Unpad() bool {
	return c.unpad
}

func (c stubCipher) InputSize(size int) int {
	if c.inputSize != nil {
		return c.inputSize(size)
	}
	return size
}

func (c stubCipher) OutputSize(size int) int {
	if c.outputSize != nil {
		return c.outputSize(size)
	}
	return size
}

type stubPadding struct {
	padFn   func([]byte, int) error
	unpadFn func([]byte) ([]byte, error)
}

func (p stubPadding) Pad(buf []byte, ori int) error {
	if p.padFn != nil {
		return p.padFn(buf, ori)
	}
	return nil
}

func (p stubPadding) Unpad(padded []byte) ([]byte, error) {
	if p.unpadFn != nil {
		return p.unpadFn(padded)
	}
	return padded, nil
}

var errTest = errors.New("test error")

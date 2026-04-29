package transport

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
)

func newPipeTransceivers(t *testing.T, clientSync, serverSync ISynchronizer) (*Transceiver, *Transceiver) {
	t.Helper()

	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() {
		clientConn.Close()
		serverConn.Close()
	})

	client := &Transceiver{
		Conn:         clientConn,
		Encoder:      codec.NewEncoder(),
		Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
		Timeout:      time.Second,
		Synchronizer: clientSync,
	}
	server := &Transceiver{
		Conn:         serverConn,
		Encoder:      codec.NewEncoder(),
		Decoder:      codec.NewDecoder(gtp.DefaultMsgCreator()),
		Timeout:      time.Second,
		Synchronizer: serverSync,
	}

	t.Cleanup(func() {
		client.Dispose()
		server.Dispose()
	})

	return client, server
}

func newUnsequencedPipeTransceivers(t *testing.T) (*Transceiver, *Transceiver) {
	t.Helper()
	return newPipeTransceivers(t, NewUnsequencedSynchronizer(), NewUnsequencedSynchronizer())
}

func newPayloadEvent(data string) IEvent {
	return Event[*gtp.MsgPayload]{
		Msg: &gtp.MsgPayload{Data: []byte(data)},
	}.Interface()
}

func recvWithTimeout(t *testing.T, tr *Transceiver) IEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	e, err := tr.Recv(ctx)
	if err != nil {
		t.Fatalf("Recv failed: %v", err)
	}
	return e
}

type stubConn struct {
	readFn             func([]byte) (int, error)
	writeFn            func([]byte) (int, error)
	closeFn            func() error
	setDeadlineFn      func(time.Time) error
	setReadDeadlineFn  func(time.Time) error
	setWriteDeadlineFn func(time.Time) error
}

func (c *stubConn) Read(p []byte) (int, error) {
	if c.readFn != nil {
		return c.readFn(p)
	}
	return 0, io.EOF
}

func (c *stubConn) Write(p []byte) (int, error) {
	if c.writeFn != nil {
		return c.writeFn(p)
	}
	return len(p), nil
}

func (c *stubConn) Close() error {
	if c.closeFn != nil {
		return c.closeFn()
	}
	return nil
}

func (c *stubConn) LocalAddr() net.Addr  { return stubAddr("local") }
func (c *stubConn) RemoteAddr() net.Addr { return stubAddr("remote") }

func (c *stubConn) SetDeadline(t time.Time) error {
	if c.setDeadlineFn != nil {
		return c.setDeadlineFn(t)
	}
	return nil
}

func (c *stubConn) SetReadDeadline(t time.Time) error {
	if c.setReadDeadlineFn != nil {
		return c.setReadDeadlineFn(t)
	}
	return nil
}

func (c *stubConn) SetWriteDeadline(t time.Time) error {
	if c.setWriteDeadlineFn != nil {
		return c.setWriteDeadlineFn(t)
	}
	return nil
}

type stubAddr string

func (a stubAddr) Network() string { return string(a) }
func (a stubAddr) String() string  { return string(a) }

type stubSynchronizer struct {
	writeFn    func([]byte) (int, error)
	writeToFn  func(io.Writer) (int64, error)
	validateFn func(gtp.MsgHead, []byte) error
	syncFn     func(uint32) error
	ackFn      func(uint32)
	sendSeqFn  func() uint32
	recvSeqFn  func() uint32
	ackSeqFn   func() uint32
	capFn      func() int
	cachedFn   func() int
	disposeFn  func()
}

func (s stubSynchronizer) Write(p []byte) (int, error) {
	if s.writeFn != nil {
		return s.writeFn(p)
	}
	return len(p), nil
}

func (s stubSynchronizer) WriteTo(w io.Writer) (int64, error) {
	if s.writeToFn != nil {
		return s.writeToFn(w)
	}
	return 0, nil
}

func (s stubSynchronizer) Validate(h gtp.MsgHead, buf []byte) error {
	if s.validateFn != nil {
		return s.validateFn(h, buf)
	}
	return nil
}

func (s stubSynchronizer) Synchronize(remoteRecvSeq uint32) error {
	if s.syncFn != nil {
		return s.syncFn(remoteRecvSeq)
	}
	return nil
}

func (s stubSynchronizer) Ack(ack uint32) {
	if s.ackFn != nil {
		s.ackFn(ack)
	}
}

func (s stubSynchronizer) SendSeq() uint32 {
	if s.sendSeqFn != nil {
		return s.sendSeqFn()
	}
	return 0
}

func (s stubSynchronizer) RecvSeq() uint32 {
	if s.recvSeqFn != nil {
		return s.recvSeqFn()
	}
	return 0
}

func (s stubSynchronizer) AckSeq() uint32 {
	if s.ackSeqFn != nil {
		return s.ackSeqFn()
	}
	return 0
}

func (s stubSynchronizer) Cap() int {
	if s.capFn != nil {
		return s.capFn()
	}
	return 0
}

func (s stubSynchronizer) Cached() int {
	if s.cachedFn != nil {
		return s.cachedFn()
	}
	return 0
}

func (s stubSynchronizer) Dispose() {
	if s.disposeFn != nil {
		s.disposeFn()
	}
}

func encodePacket(t *testing.T, e IEvent) []byte {
	t.Helper()
	buf, err := codec.NewEncoder().Encode(e.Flags, e.Msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	defer buf.Release()
	return bytes.Clone(buf.Payload())
}

func decodePacketHeads(t *testing.T, data []byte) []gtp.MsgHead {
	t.Helper()

	decoder := codec.NewDecoder(gtp.DefaultMsgCreator())
	var heads []gtp.MsgHead

	for offset := 0; offset < len(data); {
		mp, n, err := decoder.Decode(data[offset:], nil)
		if err != nil {
			t.Fatalf("Decode failed at offset %d: %v", offset, err)
		}
		heads = append(heads, mp.Head)
		offset += n
	}

	return heads
}

type readSequenceConn struct {
	mu    sync.Mutex
	reads []readResult
}

type readResult struct {
	data []byte
	err  error
}

func (c *readSequenceConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.reads) == 0 {
		return 0, io.EOF
	}
	r := c.reads[0]
	c.reads = c.reads[1:]
	if len(r.data) > 0 {
		n := copy(p, r.data)
		return n, r.err
	}
	return 0, r.err
}

func (c *readSequenceConn) Write(p []byte) (int, error) { return len(p), nil }
func (c *readSequenceConn) Close() error                { return nil }
func (c *readSequenceConn) LocalAddr() net.Addr         { return stubAddr("local") }
func (c *readSequenceConn) RemoteAddr() net.Addr        { return stubAddr("remote") }
func (c *readSequenceConn) SetDeadline(time.Time) error { return nil }
func (c *readSequenceConn) SetReadDeadline(time.Time) error {
	return nil
}
func (c *readSequenceConn) SetWriteDeadline(time.Time) error {
	return nil
}

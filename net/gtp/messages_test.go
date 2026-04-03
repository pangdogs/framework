package gtp

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"git.golaxy.org/framework/utils/binaryutil"
)

func TestBuiltinMessagesMarshalUnmarshalAndClone(t *testing.T) {
	msgs := []Msg{
		&MsgHello{
			Version:   Version_V1_0,
			SessionId: "session-1",
			Random:    []byte{1, 2, 3, 4},
			CipherSuite: CipherSuite{
				SecretKeyExchange:   SecretKeyExchange_ECDHE,
				SymmetricEncryption: SymmetricEncryption_AES,
				BlockCipherMode:     BlockCipherMode_CBC,
				PaddingMode:         PaddingMode_Pkcs7,
				HMAC:                Hash_SHA256,
			},
			Compression: Compression_Gzip,
		},
		&MsgECDHESecretKeyExchange{
			NamedCurve: NamedCurve_P256,
			PublicKey:  []byte{1, 2, 3},
			IV:         []byte{4, 5, 6},
			Nonce:      []byte{7, 8},
			NonceStep:  []byte{9},
			SignatureAlgorithm: SignatureAlgorithm{
				AsymmetricEncryption: AsymmetricEncryption_ECDSA,
				PaddingMode:          PaddingMode_None,
				Hash:                 Hash_SHA256,
			},
			Signature: []byte{10, 11, 12},
		},
		&MsgChangeCipherSpec{EncryptedHello: []byte{13, 14, 15}},
		&MsgAuth{UserId: "user", Token: "token", Extensions: []byte{16, 17}},
		&MsgContinue{SendSeq: 18, RecvSeq: 19},
		&MsgFinished{SendSeq: 20, RecvSeq: 21},
		&MsgRst{Code: Code_Reject, Message: "denied"},
		&MsgHeartbeat{},
		&MsgSyncTime{CorrId: 22, LocalTime: 23, RemoteTime: 24},
		&MsgPayload{Data: []byte("payload")},
	}

	for _, msg := range msgs {
		t.Run(reflect.TypeOf(msg).Elem().Name(), func(t *testing.T) {
			testMsgRoundTrip(t, msg)
			testMsgClone(t, msg)
		})
	}
}

func TestCipherSuiteAndSignatureAlgorithmReadWrite(t *testing.T) {
	cs := CipherSuite{
		SecretKeyExchange:   SecretKeyExchange_ECDHE,
		SymmetricEncryption: SymmetricEncryption_AES,
		BlockCipherMode:     BlockCipherMode_CFB,
		PaddingMode:         PaddingMode_X923,
		HMAC:                Hash_SHA512,
	}
	buf := make([]byte, cs.Size())
	if n, err := cs.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("CipherSuite.Read = (%d, %v)", n, err)
	}

	var decoded CipherSuite
	if n, err := decoded.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("CipherSuite.Write = (%d, %v)", n, err)
	}
	if decoded != cs {
		t.Fatalf("unexpected cipher suite round trip: %+v", decoded)
	}

	sa := SignatureAlgorithm{
		AsymmetricEncryption: AsymmetricEncryption_RSA,
		PaddingMode:          PaddingMode_PSS,
		Hash:                 Hash_SHA384,
	}
	buf = make([]byte, sa.Size())
	if n, err := sa.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("SignatureAlgorithm.Read = (%d, %v)", n, err)
	}

	var decodedSA SignatureAlgorithm
	if n, err := decodedSA.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("SignatureAlgorithm.Write = (%d, %v)", n, err)
	}
	if decodedSA != sa {
		t.Fatalf("unexpected signature algorithm round trip: %+v", decodedSA)
	}
}

func TestMsgHeadReadWriteAndSize(t *testing.T) {
	head := MsgHead{
		Len:   128,
		MsgId: MsgId_Payload,
		Flags: Flags_None().Setd(Flag_Encrypted, true).Setd(Flag_Compressed, true),
		Seq:   33,
		Ack:   22,
	}

	buf := make([]byte, head.Size())
	if n, err := head.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("MsgHead.Read = (%d, %v)", n, err)
	}

	var decoded MsgHead
	if n, err := decoded.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("MsgHead.Write = (%d, %v)", n, err)
	}
	if decoded != head {
		t.Fatalf("unexpected head round trip: %+v", decoded)
	}
}

func TestMsgPacketAndMsgPacketLen(t *testing.T) {
	packetLen := MsgPacketLen{Len: 77}
	buf := make([]byte, packetLen.Size())
	if n, err := packetLen.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("MsgPacketLen.Read = (%d, %v)", n, err)
	}

	var decodedLen MsgPacketLen
	if n, err := decodedLen.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("MsgPacketLen.Write = (%d, %v)", n, err)
	}
	if decodedLen.Len != packetLen.Len {
		t.Fatalf("unexpected packet len round trip: %d", decodedLen.Len)
	}

	msg := &MsgPayload{Data: []byte("packet")}
	packet := MsgPacket{
		Head: MsgHead{Len: uint32(msg.Size()), MsgId: msg.MsgId(), Flags: Flags_None(), Seq: 2, Ack: 1},
		Msg:  msg,
	}
	if packet.Size() != packet.Head.Size()+msg.Size() {
		t.Fatalf("unexpected packet size: %d", packet.Size())
	}

	payload := make([]byte, packet.Size())
	if n, err := packet.Read(payload); err != io.EOF || n != len(payload) {
		t.Fatalf("MsgPacket.Read = (%d, %v)", n, err)
	}

	var decodedHead MsgHead
	headSize := decodedHead.Size()
	if _, err := decodedHead.Write(payload[:headSize]); err != nil {
		t.Fatalf("decode head failed: %v", err)
	}
	if decodedHead != packet.Head {
		t.Fatalf("unexpected packet head: %+v", decodedHead)
	}

	var decodedMsg MsgPayload
	if err := Unmarshal(&decodedMsg, payload[headSize:]); err != nil {
		t.Fatalf("decode packet payload failed: %v", err)
	}
	if !bytes.Equal(decodedMsg.Data, msg.Data) {
		t.Fatalf("unexpected packet payload: %q", decodedMsg.Data)
	}

	headOnly := MsgPacket{Head: packet.Head}
	buf = make([]byte, headOnly.Head.Size())
	if n, err := headOnly.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("head-only MsgPacket.Read = (%d, %v)", n, err)
	}
}

func TestMsgSignedAndMsgCompressed(t *testing.T) {
	signed := MsgSigned{Data: []byte("abc"), MAC: []byte("mac")}
	buf := make([]byte, signed.Size())
	if n, err := signed.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("MsgSigned.Read = (%d, %v)", n, err)
	}
	var decodedSigned MsgSigned
	if n, err := decodedSigned.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("MsgSigned.Write = (%d, %v)", n, err)
	}
	if !bytes.Equal(decodedSigned.Data, signed.Data) || !bytes.Equal(decodedSigned.MAC, signed.MAC) {
		t.Fatalf("unexpected MsgSigned round trip: %+v", decodedSigned)
	}

	compressed := MsgCompressed{Data: []byte("zip"), OriginalSize: 42}
	buf = make([]byte, compressed.Size())
	if n, err := compressed.Read(buf); err != io.EOF || n != len(buf) {
		t.Fatalf("MsgCompressed.Read = (%d, %v)", n, err)
	}
	var decodedCompressed MsgCompressed
	if n, err := decodedCompressed.Write(buf); err != nil || n != len(buf) {
		t.Fatalf("MsgCompressed.Write = (%d, %v)", n, err)
	}
	if !bytes.Equal(decodedCompressed.Data, compressed.Data) || decodedCompressed.OriginalSize != compressed.OriginalSize {
		t.Fatalf("unexpected MsgCompressed round trip: %+v", decodedCompressed)
	}
}

func TestMarshalAndUnmarshalErrors(t *testing.T) {
	if _, err := Marshal[*brokenReadableMsg](&brokenReadableMsg{}); err == nil {
		t.Fatal("expected Marshal error")
	}
	if err := Unmarshal(&brokenWritableMsg{}, []byte{1, 2, 3}); err == nil {
		t.Fatal("expected Unmarshal error")
	}
}

type brokenReadableMsg struct{}

func (*brokenReadableMsg) Read([]byte) (int, error)  { return 0, io.ErrUnexpectedEOF }
func (*brokenReadableMsg) Write([]byte) (int, error) { return 0, nil }
func (*brokenReadableMsg) Size() int                 { return binaryutil.SizeofUint8 }
func (*brokenReadableMsg) MsgId() MsgId              { return MsgId_Customize }
func (m *brokenReadableMsg) Clone() Msg              { return &brokenReadableMsg{} }

type brokenWritableMsg struct{}

func (*brokenWritableMsg) Read([]byte) (int, error)  { return 0, io.EOF }
func (*brokenWritableMsg) Write([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (*brokenWritableMsg) Size() int                 { return 0 }
func (*brokenWritableMsg) MsgId() MsgId              { return MsgId_Customize }
func (m *brokenWritableMsg) Clone() Msg              { return &brokenWritableMsg{} }

func testMsgRoundTrip(t *testing.T, msg Msg) {
	t.Helper()

	data, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	defer data.Release()

	creator := NewMsgCreator()
	creator.Declare(msg.Clone())

	got, err := creator.New(msg.MsgId())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := Unmarshal(got, data.Payload()); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Fatalf("unexpected round trip: got %#v want %#v", got, msg)
	}
}

func testMsgClone(t *testing.T, msg Msg) {
	t.Helper()

	clone := msg.Clone()
	if reflect.TypeOf(clone) != reflect.TypeOf(msg) {
		t.Fatalf("unexpected clone type: %T", clone)
	}
	if !reflect.DeepEqual(clone, msg) {
		t.Fatalf("unexpected clone value: got %#v want %#v", clone, msg)
	}

	switch original := msg.(type) {
	case *MsgHello:
		cloned := clone.(*MsgHello)
		original.Random[0] ^= 0xff
		if bytes.Equal(original.Random, cloned.Random) {
			t.Fatal("expected MsgHello clone to deep copy Random")
		}
	case *MsgECDHESecretKeyExchange:
		cloned := clone.(*MsgECDHESecretKeyExchange)
		original.PublicKey[0] ^= 0xff
		if bytes.Equal(original.PublicKey, cloned.PublicKey) {
			t.Fatal("expected MsgECDHESecretKeyExchange clone to deep copy PublicKey")
		}
	case *MsgChangeCipherSpec:
		cloned := clone.(*MsgChangeCipherSpec)
		original.EncryptedHello[0] ^= 0xff
		if bytes.Equal(original.EncryptedHello, cloned.EncryptedHello) {
			t.Fatal("expected MsgChangeCipherSpec clone to deep copy EncryptedHello")
		}
	case *MsgAuth:
		cloned := clone.(*MsgAuth)
		original.Extensions[0] ^= 0xff
		if bytes.Equal(original.Extensions, cloned.Extensions) {
			t.Fatal("expected MsgAuth clone to deep copy Extensions")
		}
	case *MsgPayload:
		cloned := clone.(*MsgPayload)
		original.Data[0] ^= 0xff
		if bytes.Equal(original.Data, cloned.Data) {
			t.Fatal("expected MsgPayload clone to deep copy Data")
		}
	}
}

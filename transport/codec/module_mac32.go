package codec

import (
	"errors"
	"hash"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/binaryutil"
)

// MAC32Module MAC32模块
type MAC32Module struct {
	Hash       hash.Hash32 // hash(32bit)函数
	PrivateKey []byte      // 秘钥
	gcList     [][]byte    // GC列表
}

// PatchMAC 补充MAC
func (m *MAC32Module) PatchMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, errors.New("setting Hash is nil")
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgBuf)
	m.Hash.Write(m.PrivateKey)

	msgMAC := transport.MsgMAC32{
		Data: msgBuf,
		MAC:  m.Hash.Sum32(),
	}

	buf := BytesPool.Get(msgMAC.Size())
	defer func() {
		if err == nil {
			m.gcList = append(m.gcList, buf)
		} else {
			BytesPool.Put(buf)
		}
	}()

	_, err = msgMAC.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// VerifyMAC 验证MAC
func (m *MAC32Module) VerifyMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, errors.New("setting Hash is nil")
	}

	msgMAC := transport.MsgMAC32{}

	_, err = msgMAC.Write(msgBuf)
	if err != nil {
		return nil, err
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgMAC.Data)
	m.Hash.Write(m.PrivateKey)

	if m.Hash.Sum32() != msgMAC.MAC {
		return nil, ErrIncorrectMAC
	}

	return msgMAC.Data, nil
}

// SizeofMAC MAC大小
func (m *MAC32Module) SizeofMAC(msgLen int) int {
	return binaryutil.SizeofVarint(int64(msgLen)) + binaryutil.SizeofUint32()
}

// GC GC
func (m *MAC32Module) GC() {
	for i := range m.gcList {
		BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

package codec

import (
	"bytes"
	"errors"
	"hash"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/binaryutil"
)

var (
	ErrIncorrectMAC = errors.New("incorrect MAC")
)

// IMACModule MAC模块接口
type IMACModule interface {
	// PatchMAC 补充MAC
	PatchMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error)
	// VerifyMAC 验证MAC
	VerifyMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error)
	// SizeofMAC MAC大小
	SizeofMAC(msgLen int) int
	// GC GC
	GC()
}

// MACModule MAC模块
type MACModule struct {
	Hash       hash.Hash // hash函数
	PrivateKey []byte    // 秘钥
	gcList     [][]byte  // GC列表
}

// PatchMAC 补充MAC
func (m *MACModule) PatchMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, errors.New("setting Hash is nil")
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgBuf)
	m.Hash.Write(m.PrivateKey)

	msgMAC := transport.MsgMAC{
		Data: msgBuf,
		MAC:  m.Hash.Sum(nil),
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
func (m *MACModule) VerifyMAC(msgId transport.MsgId, flags transport.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, errors.New("setting Hash is nil")
	}

	msgMAC := transport.MsgMAC{}

	_, err = msgMAC.Write(msgBuf)
	if err != nil {
		return nil, err
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgMAC.Data)
	m.Hash.Write(m.PrivateKey)

	if bytes.Compare(m.Hash.Sum(nil), msgMAC.MAC) != 0 {
		return nil, ErrIncorrectMAC
	}

	return msgMAC.Data, nil
}

// SizeofMAC MAC大小
func (m *MACModule) SizeofMAC(msgLen int) int {
	return binaryutil.SizeofVarint(int64(msgLen)) + binaryutil.SizeofVarint(int64(m.Hash.Size())) + m.Hash.Size()
}

// GC GC
func (m *MACModule) GC() {
	for i := range m.gcList {
		BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

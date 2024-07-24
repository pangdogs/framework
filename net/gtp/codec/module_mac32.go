package codec

import (
	"errors"
	"fmt"
	"git.golaxy.org/core"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/utils/binaryutil"
	"hash"
)

// NewMAC32Module 创建MAC32模块
func NewMAC32Module(h hash.Hash32, pk []byte) IMACModule {
	if h == nil {
		panic(fmt.Errorf("%w: h is nil", core.ErrArgs))
	}

	if len(pk) <= 0 {
		panic(fmt.Errorf("%w: len(pk) <= 0", core.ErrArgs))
	}

	return &MAC32Module{
		Hash:       h,
		PrivateKey: pk,
	}
}

// MAC32Module MAC32模块
type MAC32Module struct {
	Hash       hash.Hash32 // hash(32bit)函数
	PrivateKey []byte      // 秘钥
}

// PatchMAC 补充MAC
func (m *MAC32Module) PatchMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst binaryutil.RecycleBytes, err error) {
	if m.Hash == nil {
		return binaryutil.NilRecycleBytes, errors.New("setting Hash is nil")
	}

	m.Hash.Reset()
	bs := [2]byte{msgId, byte(flags)}
	m.Hash.Write(bs[:])
	m.Hash.Write(msgBuf)
	m.Hash.Write(m.PrivateKey)

	msgMAC := gtp.MsgMAC32{
		Data: msgBuf,
		MAC:  m.Hash.Sum32(),
	}

	buf := binaryutil.MakeRecycleBytes(msgMAC.Size())
	defer func() {
		if !buf.Equal(dst) {
			buf.Release()
		}
	}()

	_, err = msgMAC.Read(buf.Data())
	if err != nil {
		return binaryutil.NilRecycleBytes, err
	}

	return buf, nil
}

// VerifyMAC 验证MAC
func (m *MAC32Module) VerifyMAC(msgId gtp.MsgId, flags gtp.Flags, msgBuf []byte) (dst []byte, err error) {
	if m.Hash == nil {
		return nil, errors.New("setting Hash is nil")
	}

	msgMAC := gtp.MsgMAC32{}

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

package codec

import (
	"errors"
	"kit.golaxy.org/golaxy/util/generic"
	"kit.golaxy.org/plugins/gtp/method"
	"kit.golaxy.org/plugins/util/binaryutil"
)

type (
	FetchNonce = generic.PairFunc0[[]byte, error] // 获取nonce值
)

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) ([]byte, error)
	// SizeOfAddition 附加数据大小
	SizeOfAddition(msgLen int) (int, error)
	// GC GC
	GC()
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	Cipher     method.Cipher  // 对称密码算法
	Padding    method.Padding // 填充方案
	FetchNonce FetchNonce     // 获取nonce值
	gcList     [][]byte       // GC列表
}

// Transforming 变换数据
func (m *EncryptionModule) Transforming(dst, src []byte) (ret []byte, err error) {
	if m.Cipher == nil {
		return nil, errors.New("setting Cipher is nil")
	}

	var in []byte

	is := m.Cipher.InputSize(len(src))
	if is > len(src) {
		buf := binaryutil.BytesPool.Get(is)
		defer binaryutil.BytesPool.Put(buf)

		copy(buf, src)
		in = buf
	} else {
		in = src
	}

	os := m.Cipher.OutputSize(len(src))
	if os > len(dst) {
		buf := binaryutil.BytesPool.Get(os)
		defer func() {
			if err == nil {
				m.gcList = append(m.gcList, buf)
			} else {
				binaryutil.BytesPool.Put(buf)
			}
		}()

		ret = buf
	} else {
		ret = dst
	}

	if m.Cipher.Pad() {
		if m.Padding == nil {
			return nil, errors.New("setting Padding is nil")
		}
		err = m.Padding.Pad(in, len(src))
		if err != nil {
			return nil, err
		}
	}

	var nonce []byte

	if m.Cipher.NonceSize() > 0 {
		if m.FetchNonce == nil {
			return nil, errors.New("setting FetchNonce is nil")
		}
		nonce, err = generic.PairFuncError(m.FetchNonce.Invoke())
		if err != nil {
			return nil, err
		}
	}

	ts, err := m.Cipher.Transforming(ret, in, nonce)
	if err != nil {
		return nil, err
	}
	ret = ret[:ts]

	if m.Cipher.Unpad() {
		if m.Padding == nil {
			return nil, errors.New("setting Padding is nil")
		}
		ret, err = m.Padding.Unpad(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// SizeOfAddition 附加数据大小
func (m *EncryptionModule) SizeOfAddition(msgLen int) (int, error) {
	if m.Cipher == nil {
		return 0, errors.New("setting Cipher is nil")
	}
	size := m.Cipher.OutputSize(msgLen) - msgLen
	if size < 0 {
		return 0, nil
	}
	return size, nil
}

// GC GC
func (m *EncryptionModule) GC() {
	for i := range m.gcList {
		binaryutil.BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

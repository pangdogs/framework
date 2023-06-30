package codec

import (
	"errors"
	"kit.golaxy.org/plugins/transport/method"
)

type (
	FetchNonce = func() ([]byte, error)
)

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) ([]byte, error)
	// DeltaSize 变换后数据增减大小（预估值）
	DeltaSize() (int, error)
	// GC GC
	GC()
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	CipherStream method.CipherStream // 密码流
	Padding      method.Padding      // 填充方案
	FetchNonce   FetchNonce          // 获取nonce值
	gcList       [][]byte            // GC列表
}

// Transforming 变换数据
func (m *EncryptionModule) Transforming(dst, src []byte) (ret []byte, err error) {
	if m.CipherStream == nil {
		return nil, errors.New("setting CipherStream is nil")
	}

	var in []byte

	is := m.CipherStream.InputSize(len(src))
	if is > len(src) {
		buf := BytesPool.Get(is)
		defer BytesPool.Put(buf)

		copy(buf, src)
		in = buf
	} else {
		in = src
	}

	os := m.CipherStream.OutputSize(len(src))
	if os > len(dst) {
		buf := BytesPool.Get(os)
		defer func() {
			if err == nil {
				m.gcList = append(m.gcList, buf)
			} else {
				BytesPool.Put(buf)
			}
		}()

		ret = buf
	} else {
		ret = dst
	}

	if m.CipherStream.Pad() {
		if m.Padding == nil {
			return nil, errors.New("setting Padding is nil")
		}
		err = m.Padding.Pad(in, len(src))
		if err != nil {
			return nil, err
		}
	}

	nonce, err := m.FetchNonce()
	if err != nil {
		return nil, err
	}

	ts, err := m.CipherStream.Transforming(ret, in, nonce)
	if err != nil {
		return nil, err
	}
	ret = ret[:ts]

	if m.CipherStream.Unpad() {
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

// DeltaSize 变换后数据增减大小（预估值）
func (m *EncryptionModule) DeltaSize() (int, error) {
	if m.CipherStream == nil {
		return 0, errors.New("setting CipherStream is nil")
	}
	return m.CipherStream.OutputSize(0), nil
}

// GC GC
func (m *EncryptionModule) GC() {
	for i := range m.gcList {
		BytesPool.Put(m.gcList[i])
	}
	m.gcList = m.gcList[:0]
}

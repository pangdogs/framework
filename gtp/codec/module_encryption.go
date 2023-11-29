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
	Transforming(dst, src []byte) (binaryutil.RecycleBytes, error)
	// SizeOfAddition 附加数据大小
	SizeOfAddition(msgLen int) (int, error)
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	Cipher     method.Cipher  // 对称密码算法
	Padding    method.Padding // 填充方案
	FetchNonce FetchNonce     // 获取nonce值
}

// Transforming 变换数据
func (m *EncryptionModule) Transforming(dst, src []byte) (ret binaryutil.RecycleBytes, err error) {
	if m.Cipher == nil {
		return binaryutil.MakeNonRecycleBytes(nil), errors.New("setting Cipher is nil")
	}

	var in []byte

	is := m.Cipher.InputSize(len(src))
	if is > len(src) {
		buf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(is))
		defer buf.Release()

		copy(buf.Data(), src)
		in = buf.Data()
	} else {
		in = src
	}

	os := m.Cipher.OutputSize(len(src))
	if os > len(dst) {
		buf := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(os))
		defer func() {
			if err != nil {
				buf.Release()
			}
		}()

		ret = buf
	} else {
		ret = binaryutil.MakeNonRecycleBytes(dst)
	}

	if m.Cipher.Pad() {
		if m.Padding == nil {
			return binaryutil.MakeNonRecycleBytes(nil), errors.New("setting Padding is nil")
		}
		err = m.Padding.Pad(in, len(src))
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), err
		}
	}

	var nonce []byte

	if m.Cipher.NonceSize() > 0 {
		if m.FetchNonce == nil {
			return binaryutil.MakeNonRecycleBytes(nil), errors.New("setting FetchNonce is nil")
		}
		nonce, err = generic.PairFuncError(m.FetchNonce.Invoke())
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), err
		}
	}

	ts, err := m.Cipher.Transforming(ret.Data(), in, nonce)
	if err != nil {
		return binaryutil.MakeNonRecycleBytes(nil), err
	}
	ret = binaryutil.SliceRecycleBytes(ret, 0, ts)

	if m.Cipher.Unpad() {
		if m.Padding == nil {
			return binaryutil.MakeNonRecycleBytes(nil), errors.New("setting Padding is nil")
		}
		buf, err := m.Padding.Unpad(ret.Data())
		if err != nil {
			return binaryutil.MakeNonRecycleBytes(nil), err
		}
		ret = binaryutil.SliceRecycleBytes(ret, 0, len(buf))
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

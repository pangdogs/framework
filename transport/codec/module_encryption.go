package codec

import (
	"errors"
	"kit.golaxy.org/plugins/transport/method"
)

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Transforming 变换数据
	Transforming(dst, src []byte) error
	// Parallel 可否并行执行
	Parallel() (bool, error)
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	CipherStream method.CipherStream
}

// Transforming 变换数据
func (m *EncryptionModule) Transforming(dst, src []byte) error {
	if m.CipherStream == nil {
		return errors.New("setting CipherStream is nil")
	}

	if len(dst) < len(src) {
		return errors.New("dst smaller than src")
	}

	m.CipherStream.Transforming(dst, src)

	return nil
}

// Parallel 可否并行执行
func (m *EncryptionModule) Parallel() (bool, error) {
	if m.CipherStream == nil {
		return false, errors.New("setting CipherStream is nil")
	}
	return m.CipherStream.Parallel(), nil
}

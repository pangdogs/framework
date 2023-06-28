package codec

import (
	"errors"
	"kit.golaxy.org/plugins/transport/method"
)

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Encrypt 加密数据
	Encrypt(dst, src []byte) error
	// Decrypt 解密数据
	Decrypt(dst, src []byte) error
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	Encrypter, Decrypter method.CipherStream
}

// Encrypt 加密数据
func (m *EncryptionModule) Encrypt(dst, src []byte) error {
	if m.Encrypter == nil {
		return errors.New("setting Encrypter is nil")
	}

	if len(dst) < len(src) {
		return errors.New("dst smaller than src")
	}

	m.Encrypter.Transforming(dst, src)

	return nil
}

// Decrypt 解密数据
func (m *EncryptionModule) Decrypt(dst, src []byte) error {
	if m.Decrypter == nil {
		return errors.New("setting Decrypter is nil")
	}

	if len(dst) < len(src) {
		return errors.New("dst smaller than src")
	}

	m.Decrypter.Transforming(dst, src)

	return nil
}

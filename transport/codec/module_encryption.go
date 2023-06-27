package codec

import (
	"crypto/cipher"
	"errors"
)

// IEncryptionModule 加密模块接口
type IEncryptionModule interface {
	// Transform 加解密数据
	Transform(dst, src []byte) error
}

// EncryptionModule 加密模块
type EncryptionModule struct {
	CipherStream cipher.Stream // 密码流
}

// Transform 加解密数据
func (m *EncryptionModule) Transform(dst, src []byte) error {
	if m.CipherStream == nil {
		return errors.New("setting CipherStream is nil")
	}

	if len(dst) < len(src) {
		return errors.New("dst bytes smaller than src")
	}

	m.CipherStream.XORKeyStream(dst, src)

	return nil
}

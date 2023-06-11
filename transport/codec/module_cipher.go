package codec

import (
	"crypto/cipher"
	"errors"
)

// ICipherModule 加密模块接口
type ICipherModule interface {
	// XORKeyStream 密码流转换
	XORKeyStream(dst, src []byte) error
}

// CipherModule 加密模块
type CipherModule struct {
	StreamCipher cipher.Stream // 密码流
}

// XORKeyStream 密码流转换
func (m *CipherModule) XORKeyStream(dst, src []byte) error {
	if m.StreamCipher == nil {
		return errors.New("setting StreamCipher is nil")
	}

	if len(dst) < len(src) {
		return errors.New("dst bytes smaller than src")
	}

	m.StreamCipher.XORKeyStream(dst, src)

	return nil
}

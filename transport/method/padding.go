package method

import (
	"errors"
	"kit.golaxy.org/plugins/transport"
)

// Padding 填充方案
type Padding interface {
	// Pad 填充
	Pad(buf []byte, ori int) error
	// Unpad 解除填充
	Unpad(padded []byte) ([]byte, error)
}

// NewPadding 创建填充方案
func NewPadding(pm transport.PaddingMode) (Padding, error) {
	switch pm {
	case transport.PaddingMode_Pkcs7:
		return _Pkcs7{}, nil
	case transport.PaddingMode_X923:
		return _X923{}, nil
	default:
		return nil, ErrInvalidMethod
	}
}

type _Pkcs7 struct{}

// Pad 填充
func (_Pkcs7) Pad(buf []byte, ori int) error {
	padLen := len(buf) - ori
	if padLen <= 0 || padLen > 0xff {
		return errors.New("pkcs7: wrong pad length")
	}
	for i := 0; i < padLen; i++ {
		buf[ori+i] = byte(padLen)
	}
	return nil
}

// Unpad 解除填充
func (_Pkcs7) Unpad(padded []byte) ([]byte, error) {
	padLen := padded[len(padded)-1]
	padPos := len(padded) - int(padLen)
	if padPos < 0 {
		return nil, errors.New("pkcs7: wrong pad pos")
	}

	for i := len(padded) - 1; i >= padPos; i-- {
		if padded[i] != padLen {
			return nil, errors.New("pkcs7: incorrect padded")
		}
	}

	return padded[:padPos], nil
}

type _X923 struct{}

// Pad 填充
func (_X923) Pad(buf []byte, ori int) error {
	padLen := len(buf) - ori
	if padLen <= 0 || padLen > 0xff {
		return errors.New("x923: wrong pad length")
	}
	for i := 0; i < padLen-1; i++ {
		buf[ori+i] = 0
	}
	buf[ori+padLen-1] = byte(padLen)
	return nil
}

// Unpad 解除填充
func (_X923) Unpad(padded []byte) ([]byte, error) {
	padLen := padded[len(padded)-1]
	padPos := len(padded) - int(padLen)
	if padPos < 0 {
		return nil, errors.New("x923: wrong pad pos")
	}

	for i := len(padded) - 2; i >= padPos; i-- {
		if padded[i] != 0 {
			return nil, errors.New("x923: incorrect padded")
		}
	}

	return padded[:padPos], nil
}

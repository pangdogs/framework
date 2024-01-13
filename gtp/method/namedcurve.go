package method

import (
	"crypto/ecdh"
	"git.golaxy.org/plugins/gtp"
)

// NewNamedCurve 创建曲线
func NewNamedCurve(nc gtp.NamedCurve) (ecdh.Curve, error) {
	switch nc {
	case gtp.NamedCurve_X25519:
		return ecdh.X25519(), nil
	case gtp.NamedCurve_P256:
		return ecdh.P256(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

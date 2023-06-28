package method

import (
	"crypto/ecdh"
	"kit.golaxy.org/plugins/transport"
)

// NewNamedCurve 创建曲线
func NewNamedCurve(nc transport.NamedCurve) (ecdh.Curve, error) {
	switch nc {
	case transport.NamedCurve_X25519:
		return ecdh.X25519(), nil
	case transport.NamedCurve_Secp256r1:
		return ecdh.P256(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

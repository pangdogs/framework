package method

import (
	"crypto/ecdh"
	"kit.golaxy.org/plugins/transport"
)

func NewNamedCurve(m transport.NamedCurve) (ecdh.Curve, error) {
	switch m {
	case transport.NamedCurve_X25519:
		return ecdh.X25519(), nil
	case transport.NamedCurve_Secp256r1:
		return ecdh.P256(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

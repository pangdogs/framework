package method

import (
	"testing"

	"git.golaxy.org/framework/net/gtp"
)

func TestNewNamedCurve(t *testing.T) {
	cases := []gtp.NamedCurve{
		gtp.NamedCurve_X25519,
		gtp.NamedCurve_P256,
		gtp.NamedCurve_P384,
		gtp.NamedCurve_P521,
	}

	for _, tc := range cases {
		t.Run(tc.String(), func(t *testing.T) {
			curve, err := NewNamedCurve(tc)
			if err != nil {
				t.Fatalf("NewNamedCurve failed: %v", err)
			}
			if curve == nil {
				t.Fatal("expected curve")
			}
		})
	}
}

func TestNewNamedCurveInvalid(t *testing.T) {
	if _, err := NewNamedCurve(gtp.NamedCurve(255)); err != ErrInvalidMethod {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
}

package method

import (
	"crypto/sha256"
	"hash"
	"hash/fnv"
	"kit.golaxy.org/plugins/transport"
)

func NewHash(m transport.HashMethod) (hash.Hash, error) {
	switch m {
	case transport.HashMethod_Fnv1a128:
		return fnv.New128a(), nil
	case transport.HashMethod_SHA256:
		return sha256.New(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

func NewHash32(m transport.HashMethod) (hash.Hash32, error) {
	switch m {
	case transport.HashMethod_Fnv1a32:
		return fnv.New32a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

func NewHash64(m transport.HashMethod) (hash.Hash64, error) {
	switch m {
	case transport.HashMethod_Fnv1a64:
		return fnv.New64a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

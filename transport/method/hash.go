package method

import (
	"crypto/sha256"
	"hash"
	"hash/fnv"
	"kit.golaxy.org/plugins/transport"
)

// NewHash 创建Hash
func NewHash(h transport.Hash) (hash.Hash, error) {
	switch h {
	case transport.Hash_Fnv1a128:
		return fnv.New128a(), nil
	case transport.Hash_SHA256:
		return sha256.New(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

// NewHash32 创建Hash32
func NewHash32(h transport.Hash) (hash.Hash32, error) {
	switch h {
	case transport.Hash_Fnv1a32:
		return fnv.New32a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

// NewHash64 创建Hash64
func NewHash64(h transport.Hash) (hash.Hash64, error) {
	switch h {
	case transport.Hash_Fnv1a64:
		return fnv.New64a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

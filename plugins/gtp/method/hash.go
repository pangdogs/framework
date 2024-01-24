package method

import (
	"crypto/sha256"
	"git.golaxy.org/framework/plugins/gtp"
	"hash"
	"hash/fnv"
)

// NewHash 创建Hash
func NewHash(h gtp.Hash) (hash.Hash, error) {
	switch h {
	case gtp.Hash_Fnv1a128:
		return fnv.New128a(), nil
	case gtp.Hash_SHA256:
		return sha256.New(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

// NewHash32 创建Hash32
func NewHash32(h gtp.Hash) (hash.Hash32, error) {
	switch h {
	case gtp.Hash_Fnv1a32:
		return fnv.New32a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

// NewHash64 创建Hash64
func NewHash64(h gtp.Hash) (hash.Hash64, error) {
	switch h {
	case gtp.Hash_Fnv1a64:
		return fnv.New64a(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package method

import (
	"crypto/sha256"
	"git.golaxy.org/framework/net/gtp"
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

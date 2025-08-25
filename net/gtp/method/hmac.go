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
	"crypto/hmac"
	"crypto/sha256"
	"git.golaxy.org/framework/net/gtp"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"hash"
)

// NewHMAC 创建HMAC
func NewHMAC(h gtp.Hash, key []byte) (hash.Hash, error) {
	switch h {
	case gtp.Hash_SHA256:
		return hmac.New(sha256.New, key), nil
	case gtp.Hash_BLAKE2b256:
		return blake2b.New256(key)
	case gtp.Hash_BLAKE2b384:
		return blake2b.New384(key)
	case gtp.Hash_BLAKE2b512:
		return blake2b.New512(key)
	case gtp.Hash_BLAKE2s256:
		return blake2s.New256(key)
	default:
		return nil, ErrInvalidMethod
	}
}

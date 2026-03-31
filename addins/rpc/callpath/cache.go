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

package callpath

import (
	"hash/fnv"
	"maps"
	"runtime"
	"sync/atomic"

	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/types"
)

type _Cached struct {
	Script string
	Method string
}

var (
	cache atomic.Pointer[map[uint32]*_Cached]
)

// Cache using Hash-based Transmission to reduce network transmission overhead
func Cache(script, method string) uint32 {
	hash := fnv.New32a()

	sep := [1]byte{0}
	hash.Write(types.String2Bytes(script))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(method))

	idx := hash.Sum32()

	for {
		old := cache.Load()
		if old != nil {
			if exists, ok := (*old)[idx]; ok {
				if exists.Script == script && exists.Method == method {
					return idx
				}
				exception.Panicf("rpc: cached index %d conflict: existing %+v vs new %+v; rename the script or method to change the generated call path id", idx, *exists, _Cached{Script: script, Method: method})
			}
		}

		cached := &_Cached{
			Script: script,
			Method: method,
		}

		var next map[uint32]*_Cached
		if old != nil {
			next = maps.Clone(*old)
		} else {
			next = make(map[uint32]*_Cached)
		}

		next[idx] = cached

		if cache.CompareAndSwap(old, &next) {
			return idx
		}

		runtime.Gosched()
	}
}

func reduce(script, method string) uint32 {
	hash := fnv.New32a()

	sep := [1]byte{0}
	hash.Write(types.String2Bytes(script))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(method))

	return hash.Sum32()
}

func inflate(idx uint32) *_Cached {
	m := cache.Load()
	if m == nil {
		return nil
	}
	return (*m)[idx]
}

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
	"fmt"
	"git.golaxy.org/core/utils/types"
	"hash/fnv"
	"sync"
)

type _Cached struct {
	Plugin    string
	Component string
	Procedure string
	Method    string
}

var (
	mutex sync.RWMutex
	cache = map[uint32]*_Cached{}
)

// Cache using Hash-based Transmission to reduce network transmission overhead
func Cache(plugin, component, procedure, method string) uint32 {
	hash := fnv.New32a()

	sep := [1]byte{0}
	hash.Write(types.String2Bytes(plugin))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(component))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(procedure))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(method))

	idx := hash.Sum32()

	mutex.Lock()
	defer mutex.Unlock()

	cached := &_Cached{
		Plugin:    plugin,
		Component: component,
		Procedure: procedure,
		Method:    method,
	}

	if exists, ok := cache[idx]; ok {
		if *exists == *cached {
			return idx
		}
		panic(fmt.Errorf("cached index %d conflict: existing %+v vs new %+v", idx, *exists, *cached))
	}

	cache[idx] = cached

	return idx
}

func reduce(plugin, component, procedure, method string) uint32 {
	hash := fnv.New32a()

	sep := [1]byte{0}
	hash.Write(types.String2Bytes(plugin))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(component))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(procedure))
	hash.Write(sep[:])
	hash.Write(types.String2Bytes(method))

	return hash.Sum32()
}

func inflate(idx uint32) *_Cached {
	mutex.RLock()
	cached, ok := cache[idx]
	mutex.RUnlock()
	if !ok {
		return nil
	}
	return cached
}

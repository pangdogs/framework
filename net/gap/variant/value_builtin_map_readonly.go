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

package variant

import (
	"cmp"
	"git.golaxy.org/core/utils/generic"
)

// MakeReadonlyMapFromGoMap 创建只读map
func MakeReadonlyMapFromGoMap[K comparable, V any](m map[K]V) (Map, error) {
	varMap := make(Map, 0, len(m))

	for k, v := range m {
		varK, err := CastReadonlyVariant(k)
		if err != nil {
			return nil, err
		}

		varV, err := CastReadonlyVariant(v)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeReadonlyMapFromSliceMap 创建只读map
func MakeReadonlyMapFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for _, kv := range m {
		varK, err := CastReadonlyVariant(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastReadonlyVariant(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeReadonlyMapFromUnorderedSliceMap 创建只读map
func MakeReadonlyMapFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for _, kv := range m {
		varK, err := CastReadonlyVariant(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastReadonlyVariant(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

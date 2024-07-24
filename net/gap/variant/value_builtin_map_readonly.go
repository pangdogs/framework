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

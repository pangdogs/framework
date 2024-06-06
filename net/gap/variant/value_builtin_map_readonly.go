package variant

import (
	"cmp"
	"git.golaxy.org/core/utils/generic"
)

// MakeMapReadonlyFromGoMap 创建只读map
func MakeMapReadonlyFromGoMap[K comparable, V any](m map[K]V) (Map, error) {
	varMap := make(Map, 0, len(m))

	for k, v := range m {
		varK, err := CastVariantReadonly(k)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariantReadonly(v)
		if err != nil {
			return nil, err
		}

		varMap.CastUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeMapReadonlyFromSliceMap 创建只读map
func MakeMapReadonlyFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for _, kv := range m {
		varK, err := CastVariantReadonly(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariantReadonly(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.CastUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeMapReadonlyFromUnorderedSliceMap 创建只读map
func MakeMapReadonlyFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (Map, error) {
	varMap := make(Map, 0, len(m))

	for _, kv := range m {
		varK, err := CastVariantReadonly(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastVariantReadonly(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.CastUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

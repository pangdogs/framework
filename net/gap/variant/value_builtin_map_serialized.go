package variant

import (
	"cmp"
	"git.golaxy.org/core/utils/generic"
)

// MakeSerializedMapFromGoMap 创建已序列化map
func MakeSerializedMapFromGoMap[K comparable, V any](m map[K]V) (ret Map, err error) {
	varMap := make(Map, 0, len(m))
	defer func() {
		if ret == nil {
			varMap.Release()
		}
	}()

	for k, v := range m {
		varK, err := CastSerializedVariant(k)
		if err != nil {
			return nil, err
		}

		varV, err := CastSerializedVariant(v)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeSerializedMapFromSliceMap 创建已序列化map
func MakeSerializedMapFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (ret Map, err error) {
	varMap := make(Map, 0, len(m))
	defer func() {
		if ret == nil {
			varMap.Release()
		}
	}()

	for _, kv := range m {
		varK, err := CastSerializedVariant(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastSerializedVariant(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

// MakeSerializedMapFromUnorderedSliceMap 创建已序列化map
func MakeSerializedMapFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (ret Map, err error) {
	varMap := make(Map, 0, len(m))
	defer func() {
		if ret == nil {
			varMap.Release()
		}
	}()

	for _, kv := range m {
		varK, err := CastSerializedVariant(kv.K)
		if err != nil {
			return nil, err
		}

		varV, err := CastSerializedVariant(kv.V)
		if err != nil {
			return nil, err
		}

		varMap.ToUnorderedSliceMap().Add(varK, varV)
	}

	return varMap, nil
}

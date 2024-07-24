package variant

import (
	"cmp"
	"git.golaxy.org/core/utils/generic"
)

// MakeMapSerializedFromGoMap 创建已序列化map
func MakeMapSerializedFromGoMap[K comparable, V any](m map[K]V) (ret Map, err error) {
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

// MakeMapSerializedFromSliceMap 创建已序列化map
func MakeMapSerializedFromSliceMap[K cmp.Ordered, V any](m generic.SliceMap[K, V]) (ret Map, err error) {
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

// MakeMapSerializedFromUnorderedSliceMap 创建已序列化map
func MakeMapSerializedFromUnorderedSliceMap[K comparable, V any](m generic.UnorderedSliceMap[K, V]) (ret Map, err error) {
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

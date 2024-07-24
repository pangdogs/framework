package variant

// MakeSerializedArray 创建已序列化array
func MakeSerializedArray[T any](arr []T) (ret Array, err error) {
	varArr := make(Array, 0, len(arr))
	defer func() {
		if ret == nil {
			varArr.Release()
		}
	}()

	for i := range arr {
		v, err := CastSerializedVariant(arr[i])
		if err != nil {
			return nil, err
		}
		varArr = append(varArr, v)
	}

	return varArr, nil
}

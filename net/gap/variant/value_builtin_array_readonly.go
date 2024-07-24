package variant

// MakeReadonlyArray 创建只读array
func MakeReadonlyArray[T any](arr []T) (Array, error) {
	varArr := make(Array, 0, len(arr))

	for i := range arr {
		v, err := CastReadonlyVariant(arr[i])
		if err != nil {
			return nil, err
		}
		varArr = append(varArr, v)
	}

	return varArr, nil
}

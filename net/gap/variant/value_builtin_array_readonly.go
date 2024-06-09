package variant

// MakeArrayReadonly 创建只读array
func MakeArrayReadonly[T any](arr []T) (Array, error) {
	varArr := make(Array, 0, len(arr))

	for i := range arr {
		v, err := CastVariantReadonly(arr[i])
		if err != nil {
			return nil, err
		}
		varArr = append(varArr, v)
	}

	return varArr, nil
}

// MakeArrayBuffReadonly 创建只读array
func MakeArrayBuffReadonly[T any](arr []T) (Array, error) {
	varArr := make(Array, 0, len(arr))
	var err error

	defer func() {
		if err != nil {
			for i := range varArr {
				it := &varArr[i]

				if it.Readonly() {
					it.ValueReadonly.Release()
				}
			}
		}
	}()

	for i := range arr {
		varRaw, err := CastVariantReadonly(arr[i])
		if err != nil {
			return nil, err
		}

		valueBuff, err := MakeValueBuff(varRaw.ValueReadonly)
		if err != nil {
			return nil, err
		}

		varBuff, err := MakeVariantReadonly(valueBuff)
		if err != nil {
			valueBuff.Release()
			return nil, err
		}

		varArr = append(varArr, varBuff)
	}

	return varArr, nil
}

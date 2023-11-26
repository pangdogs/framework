package variant

import "errors"

// Variant 可变类型
type Variant struct {
	TypeId TypeId // 类型Id
	Value  Value  // 值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
	if v.Value == nil {
		return 0, errors.New("value is nil")
	}

	var rn int

	n, err := v.TypeId.Read(p)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = v.Value.Read(p[rn:])
	rn += n
	if err != nil {
		return rn, err
	}

	return rn, nil
}

// Write implements io.Writer
func (v *Variant) Write(p []byte) (int, error) {
	var wn int

	var typeId TypeId
	n, err := typeId.Write(p)
	wn += n
	if err != nil {
		return wn, err
	}

	variant, err := VariantCreator().Spawn(typeId)
	if err != nil {
		return wn, err
	}

	n, err = variant.Value.Write(p[wn:])
	wn += n
	if err != nil {
		return wn, err
	}

	*v = variant

	return wn, nil
}

// Size 大小
func (v Variant) Size() int {
	n := v.TypeId.Size()
	if v.Value != nil {
		n += v.Value.Size()
	}
	return n
}

// MakeVariant 创建可变类型
func MakeVariant(v Value) (Variant, error) {
	if v == nil {
		return Variant{}, errors.New("v is nil")
	}
	return Variant{
		TypeId: v.Type(),
		Value:  v,
	}, nil
}

// CastVariant 转换为可变类型
func CastVariant(a any) (Variant, error) {
	switch v := a.(type) {
	case int:
		value := Int(v)
		return MakeVariant(&value)
	case int8:
		value := Int8(v)
		return MakeVariant(&value)
	case int16:
		value := Int16(v)
		return MakeVariant(&value)
	case int32:
		value := Int32(v)
		return MakeVariant(&value)
	case int64:
		value := Int64(v)
		return MakeVariant(&value)
	case uint:
		value := Uint(v)
		return MakeVariant(&value)
	case uint8:
		value := Uint8(v)
		return MakeVariant(&value)
	case uint16:
		value := Uint16(v)
		return MakeVariant(&value)
	case uint32:
		value := Uint32(v)
		return MakeVariant(&value)
	case uint64:
		value := Uint64(v)
		return MakeVariant(&value)
	case float32:
		value := Float(v)
		return MakeVariant(&value)
	case float64:
		value := Double(v)
		return MakeVariant(&value)
	case bool:
		value := Bool(v)
		return MakeVariant(&value)
	case []byte:
		value := Bytes(v)
		return MakeVariant(&value)
	case string:
		value := String(v)
		return MakeVariant(&value)
	case nil:
		value := Null{}
		return MakeVariant(&value)
	case Array:
		return MakeVariant(&v)
	case *Array:
		return MakeVariant(v)
	case Map:
		return MakeVariant(&v)
	case *Map:
		return MakeVariant(v)
	case Variant:
		return v, nil
	case Error:
		return MakeVariant(&v)
	case *Error:
		return MakeVariant(v)
	case error:
		return MakeVariant(&Error{
			Code:    -1,
			Message: v.Error(),
		})
	default:
		value, ok := a.(Value)
		if !ok {
			return Variant{}, ErrNotVariant
		}
		return MakeVariant(value)
	}
}

package variant

import (
	"errors"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/util/binaryutil"
	"reflect"
)

// Variant 可变类型
type Variant struct {
	TypeId    TypeId        // 类型Id
	Value     ValueReader   // 读取值
	Reflected reflect.Value // 反射值
}

// Read implements io.Reader
func (v Variant) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := binaryutil.ReadFrom(&bs, v.TypeId); err != nil {
		return bs.BytesWritten(), err
	}

	if v.Value == nil {
		return bs.BytesWritten(), errors.New("value is nil")
	}

	if _, err := binaryutil.ReadFrom(&bs, v.Value); err != nil {
		return bs.BytesWritten(), err
	}

	return bs.BytesWritten(), nil
}

// Write implements io.Writer
func (v *Variant) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)

	if _, err := bs.WriteTo(&v.TypeId); err != nil {
		return bs.BytesRead(), err
	}

	reflected, err := v.TypeId.NewReflected()
	if err != nil {
		return bs.BytesRead(), err
	}

	value := reflected.Interface().(Value)
	if _, err := bs.WriteTo(value); err != nil {
		return bs.BytesRead(), err
	}

	v.Value = value
	v.Reflected = reflected

	return bs.BytesRead(), nil
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
func MakeVariant(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, errors.New("v is nil")
	}
	return Variant{
		TypeId: v.TypeId(),
		Value:  v,
	}, nil
}

// CastVariant 转换为可变类型
func CastVariant(a any) (Variant, error) {
retry:
	switch v := a.(type) {
	case int:
		return MakeVariant(Int(v))
	case *int:
		return MakeVariant((*Int)(v))
	case int8:
		return MakeVariant(Int8(v))
	case *int8:
		return MakeVariant((*Int8)(v))
	case int16:
		return MakeVariant(Int16(v))
	case *int16:
		return MakeVariant((*Int16)(v))
	case int32:
		return MakeVariant(Int32(v))
	case *int32:
		return MakeVariant((*Int32)(v))
	case int64:
		return MakeVariant(Int64(v))
	case *int64:
		return MakeVariant((*Int64)(v))
	case uint:
		return MakeVariant(Uint(v))
	case *uint:
		return MakeVariant((*Uint)(v))
	case uint8:
		return MakeVariant(Uint8(v))
	case *uint8:
		return MakeVariant((*Uint8)(v))
	case uint16:
		return MakeVariant(Uint16(v))
	case *uint16:
		return MakeVariant((*Uint16)(v))
	case uint32:
		return MakeVariant(Uint32(v))
	case *uint32:
		return MakeVariant((*Uint32)(v))
	case uint64:
		return MakeVariant(Uint64(v))
	case *uint64:
		return MakeVariant((*Uint64)(v))
	case float32:
		return MakeVariant(Float(v))
	case *float32:
		return MakeVariant((*Float)(v))
	case float64:
		return MakeVariant(Double(v))
	case *float64:
		return MakeVariant((*Double)(v))
	case bool:
		return MakeVariant(Bool(v))
	case *bool:
		return MakeVariant((*Bool)(v))
	case []byte:
		return MakeVariant(Bytes(v))
	case *[]byte:
		return MakeVariant((*Bytes)(v))
	case string:
		return MakeVariant(String(v))
	case *string:
		return MakeVariant((*String)(v))
	case uid.Id:
		return MakeVariant(String(v))
	case *uid.Id:
		return MakeVariant((*String)(v))
	case nil:
		return MakeVariant(Null{})
	case Array:
		return MakeVariant(v)
	case *Array:
		return MakeVariant(v)
	case []any:
		arr, err := MakeArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case *[]any:
		arr, err := MakeArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case []reflect.Value:
		arr, err := MakeArray(v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case *[]reflect.Value:
		arr, err := MakeArray(*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(arr)
	case Map:
		return MakeVariant(v)
	case *Map:
		return MakeVariant(v)
	case map[string]any:
		m, err := MakeMap[string, any](v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case *map[string]any:
		m, err := MakeMap[string, any](*v)
		if err != nil {
			return Variant{}, err
		}
		return MakeVariant(m)
	case Variant:
		return v, nil
	case *Variant:
		return *v, nil
	case Error:
		return MakeVariant(&v)
	case *Error:
		return MakeVariant(v)
	case error:
		return MakeVariant(MakeError(v))
	case ValueReader:
		return MakeVariant(v)
	case reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrNotVariant
		}
		a = v.Interface()
		goto retry
	case *reflect.Value:
		if !v.CanInterface() {
			return Variant{}, ErrNotVariant
		}
		a = v.Interface()
		goto retry
	default:
		return Variant{}, ErrNotVariant
	}
}

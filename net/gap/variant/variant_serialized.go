package variant

import (
	"fmt"
	"git.golaxy.org/core"
)

// MakeSerializedVariant 创建已序列化可变类型
func MakeSerializedVariant(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}

	sv, err := MakeSerializedValue(v)
	if err != nil {
		return Variant{}, err
	}

	return Variant{
		TypeId:          v.TypeId(),
		SerializedValue: sv,
	}, nil
}

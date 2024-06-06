package variant

import (
	"fmt"
	"git.golaxy.org/core"
)

// MakeVariantReadonly 创建只读可变类型
func MakeVariantReadonly(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}
	return Variant{
		TypeId:        v.TypeId(),
		ValueReadonly: v,
	}, nil
}

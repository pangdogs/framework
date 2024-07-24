package variant

import (
	"fmt"
	"git.golaxy.org/core"
)

// MakeReadonlyVariant 创建只读可变类型
func MakeReadonlyVariant(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("%w: v is nil", core.ErrArgs)
	}
	return Variant{
		TypeId:        v.TypeId(),
		ReadonlyValue: v,
	}, nil
}

/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package variant

import (
	"fmt"
	"git.golaxy.org/core"
)

// MakeReadonlyVariant 创建只读可变类型
func MakeReadonlyVariant(v ValueReader) (Variant, error) {
	if v == nil {
		return Variant{}, fmt.Errorf("gap-var: %w: v is nil", core.ErrArgs)
	}
	return Variant{
		TypeId:        v.TypeId(),
		ReadonlyValue: v,
	}, nil
}

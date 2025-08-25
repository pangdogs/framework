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

package method

import (
	"crypto/ecdh"
	"git.golaxy.org/framework/net/gtp"
)

// NewNamedCurve 创建曲线
func NewNamedCurve(nc gtp.NamedCurve) (ecdh.Curve, error) {
	switch nc {
	case gtp.NamedCurve_X25519:
		return ecdh.X25519(), nil
	case gtp.NamedCurve_P256:
		return ecdh.P256(), nil
	case gtp.NamedCurve_P384:
		return ecdh.P384(), nil
	case gtp.NamedCurve_P521:
		return ecdh.P521(), nil
	default:
		return nil, ErrInvalidMethod
	}
}

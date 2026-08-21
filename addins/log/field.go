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

package log

import (
	"encoding/json"
	"fmt"

	"git.golaxy.org/core/utils/types"
	"go.uber.org/zap"
)

type lazyJSON struct {
	v any
}

func (l lazyJSON) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(l.v)
	if err != nil {
		return json.Marshal(fmt.Sprintf("json.Marshal(): %s", err.Error()))
	}
	return data, nil
}

// JSON 创建一个在日志编码时才执行 json.Marshal 的字段。
// 编码失败时字段值会替换为描述错误的 JSON 字符串。
func JSON(key string, v any) zap.Field {
	return zap.Reflect(key, lazyJSON{v: v})
}

type lazyJSONRawStringer struct {
	v fmt.Stringer
}

func (l lazyJSONRawStringer) MarshalJSON() ([]byte, error) {
	if l.v == nil {
		return []byte("null"), nil
	}
	return types.String2Bytes(l.v.String()), nil
}

// JSONRawStringer 创建一个延迟调用 String 的原始 JSON 字段。
// String 的返回值必须是有效 JSON。
func JSONRawStringer(key string, v fmt.Stringer) zap.Field {
	return zap.Reflect(key, lazyJSONRawStringer{v: v})
}

type rawJSONString struct {
	v string
}

func (r rawJSONString) MarshalJSON() ([]byte, error) {
	return types.String2Bytes(r.v), nil
}

// JSONRawString 将 v 作为未经校验的原始 JSON 写入字段。
func JSONRawString(key string, v string) zap.Field {
	return zap.Reflect(key, rawJSONString{v: v})
}

type rawJSONByteString struct {
	v []byte
}

func (r rawJSONByteString) MarshalJSON() ([]byte, error) {
	return r.v, nil
}

// JSONRawByteString 将 v 作为未经校验的原始 JSON 写入字段。
func JSONRawByteString(key string, v []byte) zap.Field {
	return zap.Reflect(key, rawJSONByteString{v: v})
}

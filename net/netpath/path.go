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

package netpath

import (
	"strings"
)

// Join 使用 sep 原样连接 elems，不清理空段或重复分隔符。
func Join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}

// Split 在最后一个 sep 处分割 path；未找到分隔符时 dir 为空、file 为原路径。
func Split(sep, path string) (dir, file string) {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return "", path
	}
	return path[:idx], path[idx+len(sep):]
}

// Root 返回 path 的第一个路径段；未找到 sep 时返回原路径。
func Root(sep, path string) string {
	idx := strings.Index(path, sep)
	if idx < 0 {
		return path
	}
	return path[:idx]
}

// Base 返回 path 的最后一个路径段；未找到 sep 时返回原路径。
func Base(sep, path string) string {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return path
	}
	return path[idx+len(sep):]
}

// Dir 返回 path 最后一个 sep 之前的部分；未找到分隔符时返回空字符串。
func Dir(sep, path string) string {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return ""
	}
	return path[:idx]
}

// InDir 报告 path 是否为 dir 的严格子路径；末尾的一个 sep 会被忽略。
// path 与 dir 相等时返回 false。
func InDir(sep, path, dir string) bool {
	path = strings.TrimSuffix(path, sep)
	dir = strings.TrimSuffix(dir, sep)

	if !strings.HasPrefix(path, dir) {
		return false
	}

	return strings.HasPrefix(path[len(dir):], sep)
}

// Equal 报告 a 与 b 是否相等；双方末尾的一个 sep 会被忽略。
func Equal(sep, a, b string) bool {
	return strings.TrimSuffix(a, sep) == strings.TrimSuffix(b, sep)
}

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

func Join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}

func Split(sep, path string) (dir, file string) {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return "", path
	}
	return path[:idx], path[idx+len(sep):]
}

func Root(sep, path string) string {
	idx := strings.Index(path, sep)
	if idx < 0 {
		return path
	}
	return path[:idx]
}

func Base(sep, path string) string {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return path
	}
	return path[idx+len(sep):]
}

func Dir(sep, path string) string {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return ""
	}
	return path[:idx]
}

func InDir(sep, path, dir string) bool {
	path = strings.TrimSuffix(path, sep)
	dir = strings.TrimSuffix(dir, sep)

	if !strings.HasPrefix(path, dir) {
		return false
	}

	return strings.HasPrefix(path[len(dir):], sep)
}

func Equal(sep, a, b string) bool {
	return strings.TrimSuffix(a, sep) == strings.TrimSuffix(b, sep)
}

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

func SameDir(sep, path, dir string) bool {
	path = strings.TrimSuffix(path, sep)
	dir = strings.TrimSuffix(dir, sep)

	if !strings.HasPrefix(path, dir) {
		return false
	}

	if len(path) <= len(dir) {
		return true
	}

	return strings.HasPrefix(path[len(dir):], sep)
}

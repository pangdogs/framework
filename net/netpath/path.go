package netpath

import (
	"strings"
)

func Path(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}

func Split(sep, path string) (dir, file string) {
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		return "", path
	}
	return path[:idx], path[idx+len(sep):]
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
		return path
	}
	return path[:idx]
}

func InDir(sep, path, dir string) bool {
	dir = strings.TrimSuffix(dir, sep)

	if !strings.HasPrefix(path, dir) || len(path) <= len(dir) {
		return false
	}

	return strings.HasPrefix(path[len(dir):], sep)
}

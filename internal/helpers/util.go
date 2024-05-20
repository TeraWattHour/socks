package helpers

import (
	"path/filepath"
)

// Location represents a location in a file. Line starts at 1, Column - at 0.
type Location struct {
	Line   int
	Column int
}

// ResolvePath resolves a path from a given path and a relative referenced path.
func ResolvePath(from string, referenced string) string {
	if filepath.IsAbs(referenced) {
		return referenced[1:]
	}
	return filepath.Join(filepath.Dir(from), referenced)
}

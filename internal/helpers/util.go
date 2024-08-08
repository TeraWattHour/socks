package helpers

import (
	"path/filepath"
)

// Location represents a location in a file. Line starts at 1, Column - at 0.
type Location struct {
	File   string
	Line   int
	Column int
	Length int
}

func (l Location) SetLength(length int) Location {
	l.Length = length
	return l
}

func (l Location) WithBase(base Location) Location {
	if base.Line == l.Line {
		l.Column += base.Column
		return l
	}
	l.Line += base.Line
	return l
}

func (l Location) PointAfter() Location {
	l.Column += l.Length
	l.Length = 1
	return l
}

// ResolvePath resolves a path from a given path and a relative referenced path.
func ResolvePath(from string, referenced string) string {
	if filepath.IsAbs(referenced) {
		return referenced[1:]
	}
	return filepath.Join(filepath.Dir(from), referenced)
}

func ApplyVariable(ctx map[string]any, key string, value any) {
	if ctx != nil && key != "" {
		ctx[key] = value
	}
}

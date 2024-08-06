package helpers

import (
	"path/filepath"
)

type File struct {
	Name    string
	Content string
}

// Location represents a location in a file. Line starts at 1, Column - at 0.
type Location struct {
	Line   int
	Column int
	Cursor int
	Length int
}

func (l Location) SetLength(length int) Location {
	l.Length = length
	return l
}

func (l Location) Combine(other Location) Location {
	l.Length = other.Cursor + other.Length - l.Cursor
	return l
}

func (l Location) FromOther() Location {
	l.Column += l.Length
	l.Cursor += l.Length
	l.Length = 1
	return l
}

func (l Location) MoveBy(amount int) Location {
	l.Cursor += amount
	l.Column += amount
	return l
}

func (l Location) PointAfter() Location {
	l.Column += l.Length
	l.Cursor += l.Length
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

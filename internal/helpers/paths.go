package helpers

import "path/filepath"

func ResolvePath(from string, referenced string) string {
	if filepath.IsAbs(referenced) {
		return referenced[1:]
	}
	return filepath.Join(filepath.Dir(from), referenced)
}

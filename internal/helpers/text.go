package helpers

import "strings"

func FixedWidth(s string, size int) string {
	if len(s) > size {
		return s[:size]
	}
	return s + strings.Repeat(" ", size-len(s))
}

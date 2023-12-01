package preprocessor

import (
	"github.com/terawatthour/socks/internal/helpers"
	"testing"
)

func TestReferencedPath(t *testing.T) {
	sets := []struct {
		base       string
		referenced string
		expected   string
	}{
		{"a/b/c.html", "d", "a/b/d"},
		{"a/b/c.html", "./d", "a/b/d"},
		{"a/b/c.html", "../d", "a/d"},
		{"a/b/c.html", "../../d", "d"},
		{"a/b/c.html", "../../../d", "../d"},
		{"a/b/c.html", "/d", "d"},
	}

	for _, set := range sets {
		resolved := helpers.ResolvePath(set.base, set.referenced)
		if resolved != set.expected {
			t.Fatalf("failed to compute path: %s", resolved)
		}
	}
}

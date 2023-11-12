package filesystem

import "testing"

func TestNewFileSystem(t *testing.T) {
	fs, err := NewFileSystem("../../test_data/basic")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(fs.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(fs.Files))
	}
}

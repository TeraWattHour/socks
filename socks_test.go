package socks

import (
	"fmt"
	"testing"
)

func TestBasicEvaluation(t *testing.T) {
	s := NewSocks()
	err := s.LoadTemplates("test_data/*.html", "test_data/**/*.html")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if err := s.PreprocessTemplates(map[string]interface{}{
		"Server": "localhost",
	}); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	s.AddGlobal("now", "2019-01-01")

	type Phrase struct {
		Content  string
		Language string
	}

	res, err := s.Run("test_data/nested.html", map[string]interface{}{
		"Phrases": []Phrase{{Content: "Hello", Language: "en"}, {Content: "Hallo", Language: "de"}},
	})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	fmt.Println("result", res)
}

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

	if err := s.PreprocessTemplates(map[string]any{
		"Server": "Socks",
	}); err != nil {
		t.Errorf("Expected no error, got %s", err)
		return
	}

	s.AddGlobal("now", "2019-01-01")

	type Phrase struct {
		Content  string
		Language string
	}

	res, err := s.ExecuteToString("test_data/nested.html", map[string]interface{}{
		"Phrases": []Phrase{{Content: "Hello", Language: "en"}, {Content: "Hallo", Language: "de"}},
		"Server":  "Socks",
	})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
		return
	}

	fmt.Println(res)
}

func TestCommentRemoval(t *testing.T) {
	s := NewSocks()
	s.LoadTemplateFromString("test.html", "keep this {# remove this#}xd")
	if err := s.PreprocessTemplates(nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	res, err := s.ExecuteToString("test.html", nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if res != "keep this xd" {
		t.Errorf("Unexpected result: %s", res)
	}
}

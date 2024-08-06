package socks

import (
	"fmt"
	"testing"
)

func TestBasicEvaluation(t *testing.T) {
	s := New(&Options{
		Sanitizer: func(s string) string {
			return s
		},
	})

	if err := s.LoadTemplates("test_data/**/*.html"); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if err := s.LoadTemplates("test_data/*.html"); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if err := s.Compile(map[string]any{
		"Server": "Socks",
	}); err != nil {
		t.Errorf("unexpected error %s", err)
		return
	}

	s.AddGlobal("now", "2019-01-01")

	type Phrase struct {
		Content  string
		Language string
	}

	res, err := s.ExecuteToString("templates/header.html", map[string]interface{}{
		"Phrases": []Phrase{{Content: "Hello", Language: "en"}, {Content: "Hallo", Language: "de"}},
		"first": []any{
			"first",
			func(num int) func(num2 int) []string {
				return func(num2 int) []string {
					return []string{fmt.Sprintf("sum is %d", num+num2)}
				}
			},
		},
		"countries": []string{"UK", "DE", "FR"},
		"resolveLanguage": func(abbreviation string) string {
			switch abbreviation {
			case "en":
				return "English"
			case "de":
				return "German"
			default:
				return "unknown language"
			}
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

package evaluator

import (
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
	"time"
)

func TestEvaluatorSimple(t *testing.T) {
	template := `<html><head><title>{{ 1123 }} {{ .Title }} {{ .Time.Format(.Title).Kok("kc") }}</title></head><body><h1>{{ .Format("najs", 123) }}</h1></body></html>`

	type Nested struct {
		Kok func(s string) string
	}

	type Time struct {
		Value  time.Time
		Format func(s string) Nested
	}

	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	eval := NewEvaluator(par)
	evaluated, err := eval.Evaluate(map[string]interface{}{
		"Title": "Hello, World!",
		"Format": func(s string, test int) string {
			return "xdddd: " + s
		},
		"Time": Time{
			Value: time.Now(),
			Format: func(s string) Nested {
				return Nested{
					Kok: func(s string) string {
						return "kok: " + s
					},
				}
			},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := "<html><head><title>1123 Hello, World! kok: kc</title></head><body><h1>xdddd: najs</h1></body></html>"
	if expected != evaluated {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

func TestEvaluatorRemove(t *testing.T) {
	template := `{% slot "najs" %} xd {% end %} {% define "remove me" %} show me {% end %}`

	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	eval := NewEvaluator(par)

	evaluated, err := eval.Evaluate(map[string]interface{}{})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := " xd   show me "
	if evaluated != expected {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

func TestEvaluatorMultiple(t *testing.T) {
	template := `{! for i, v in .Statements !} {{ v }} {! end !}`

	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	eval := NewEvaluator(par)

	evaluated, err := eval.Evaluate(map[string]interface{}{
		"Statements": []string{"najs", "najs2"},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := " najs  najs2 "
	if evaluated != expected {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}

	evaluated, err = eval.Evaluate(map[string]interface{}{
		"Statements": []string{"najs3", "najs4"},
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected = " najs3  najs4 "
	if evaluated != expected {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

func TestEvaluatorForLoop(t *testing.T) {
	template := `before{! for i, v in .elements !} {{ i }} {{ v }} {! end !}after {{ .some_other }}`

	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	eval := NewEvaluator(par)
	evaluated, err := eval.Evaluate(map[string]interface{}{
		"elements":   []string{"najs", "najs2"},
		"some_other": 132,
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := "before 0 najs  1 najs2 after 132"
	if evaluated != expected {
		t.Errorf("Expected `%s`, got `%s`", expected, evaluated)
	}
}

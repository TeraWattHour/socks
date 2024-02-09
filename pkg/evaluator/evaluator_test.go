package evaluator

import (
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

func TestEvaluator(t *testing.T) {
	sets := []struct {
		expected string
		env      map[string]interface{}
		template string
	}{{
		"najs najs najs najs2 najs2 najs najs2 najs2 ",
		map[string]any{"Iterable": []string{"najs", "najs2"}},
		`@for(v in Iterable)@for(v2 in Iterable){{ v }} {{ v2 }} @endfor@endfor`,
	}, {
		"<html><head><title>1123 Hello, World!</title></head><body><p>0: <span>najs ( 0,1)</span><span>najs2 ( 0,1)</span></p><p>1: <span>najs ( 0,1)</span><span>najs2 ( 0,1)</span></p></body></html>",
		map[string]any{"Iterable": []string{"najs", "najs2"}, "Title": "Hello, World!"},
		"<html><head><title>{{ 1123 }} {{ Title }}</title></head><body>@for(statement, i in Iterable)<p>{{ i }}: @for(nested, j in Iterable)<span>{{ nested }} ( @for(nested2, k in Iterable){{ k }}@if(k < len(Iterable)-1),@endif@endfor )</span>@endfor</p>@endfor</body></html>",
	}, {
		" Hello, World! ",
		map[string]any{"Title": "Hello, World"},
		"@if(len(Title) > 5) {{ Title + \"!\" }} @endif",
	}}

	for _, set := range sets {
		tok := tokenizer.NewTokenizer(set.template)
		if err := tok.Tokenize(); err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		par := parser.NewParser(tok)
		programs, err := par.Parse()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		eval := NewEvaluator(programs, nil)
		evaluated, err := eval.Evaluate(set.env)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if set.expected != evaluated {
			t.Errorf("expected `%s`, got `%s`", set.expected, evaluated)
		}
	}
}

package evaluator

import (
	"bytes"
	"testing"

	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

func TestLoopsAndIfs(t *testing.T) {
	template := `@if(1==1) @for(a, i in A) @if(i % 2 == 0) {{ a }} says hello! @endif @endfor @endif`
	elements, err := tokenizer.Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	programs, err := parser.Parse(elements)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	evaluated := bytes.NewBufferString("")
	eval := New(programs, nil)
	err = eval.Evaluate(evaluated, map[string]interface{}{"A": []string{"a", "b", "c"}})
	expected := `   a says hello!      c says hello!   `
	if expected != evaluated.String() {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

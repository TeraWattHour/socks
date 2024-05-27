package socks

import (
	"bytes"
	"testing"

	"github.com/terawatthour/socks/tokenizer"
)

func TestLoopsAndIfs(t *testing.T) {
	template := `@if (1==1) @for (a in A with i) @if (i % 2 == 0) {{ a }} says hello! @endif @endfor @endif`
	elements, err := tokenizer.Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	programs, err := Parse(elements)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	evaluated := bytes.NewBufferString("")
	if err := newEvaluator(programs, nil).evaluate(evaluated, map[string]interface{}{"A": []string{"a", "b", "c"}}); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := `   a says hello!      c says hello!   `
	if expected != evaluated.String() {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

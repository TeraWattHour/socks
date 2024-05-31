package socks

import (
	"bytes"
	"github.com/terawatthour/socks/internal/helpers"
	"testing"

	"github.com/terawatthour/socks/tokenizer"
)

func TestLoopsAndIfs(t *testing.T) {
	template := `@if (1==1) @for (a in A with i) @if (i % 2 == 0) {{ a }} says hello! @endif @endfor @endif`
	elements, err := tokenizer.Tokenize("debug.txt", template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	programs, err := Parse(helpers.File{"debug.txt", template}, elements)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	evaluated := bytes.NewBufferString("")
	if err := newEvaluator(helpers.File{"debug.txt", template}, programs, nil).evaluate(evaluated, map[string]interface{}{"A": []string{"a", "b", "c"}}); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := `   a says hello!      c says hello!   `
	if expected != evaluated.String() {
		t.Errorf("expected `%s`, got `%s`", expected, evaluated)
	}
}

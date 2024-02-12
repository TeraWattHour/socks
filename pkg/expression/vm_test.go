package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

type tescik struct {
	Aha []string
}

func TestVM_Run(t *testing.T) {
	elements, err := tokenizer.Tokenize("{{ i < 1 }}")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	expr, err := Parse(elements[1].(*tokenizer.Mustache).Tokens)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	compiler := NewCompiler(expr.Expr)
	compiler.Compile()

	printChunk(compiler.chunk)

	vm := NewVM(compiler.chunk)
	result, err := vm.Run(map[string]any{
		"i":      0,
		"idx":    1,
		"number": int(123),
		"test":   []string{"pirwszy", "drugi"},
		"parent": map[string]any{
			"test": func(number int) tescik {
				return tescik{Aha: []string{fmt.Sprintf("num: %d", number), "num: drugi"}}
			},
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	fmt.Println("result", result)
	// Output: true
}

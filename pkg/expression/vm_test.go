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
	tok := tokenizer.NewTokenizer("{{ 1 in range(1, 4) }}")
	tok.Tokenize()
	p := NewParser(tok.Elements[0].Tokens())
	expr, err := p.Parse()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	compiler := NewCompiler(expr)
	compiler.Compile()

	printChunk(compiler.chunk)

	vm := NewVM(compiler.chunk)
	result, err := vm.Run(map[string]any{
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

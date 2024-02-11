package expression

import (
	"testing"

	"github.com/terawatthour/socks/pkg/tokenizer"
)

func TestCompiler(t *testing.T) {
	t.Run("simple expression", func(t *testing.T) {
		elements, err := tokenizer.Tokenize("{{ test.int(int(123)) }}")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		p := newParser(elements[1].(*tokenizer.Mustache).Tokens)
		expr, err := p.parser()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		compiler := NewCompiler(expr)
		compiler.Compile()

		//
		//want := Chunk{
		//	Instructions: []int{
		//		OpTrue,
		//		OpFalse,
		//		OpAnd,
		//	},
		//	Constants: []interface{}{},
		//}

		printChunk(compiler.chunk)
	})
}

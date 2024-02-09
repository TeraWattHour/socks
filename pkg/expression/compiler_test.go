package expression

import (
	"testing"

	"github.com/terawatthour/socks/pkg/tokenizer"
)

func TestCompiler(t *testing.T) {
	t.Run("simple expression", func(t *testing.T) {
		tok := tokenizer.NewTokenizer("{{ test.int(int(123)) }}")
		tok.Tokenize()
		p := NewParser(tok.Elements[0].Tokens())
		expr, err := p.Parse()
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

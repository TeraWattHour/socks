package expression

//func TestCompiler(t *testing.T) {
//	t.Run("simple expression", func(t *testing.T) {
//		elements, err := tokenizer.Tokenize("{{ 2 + [1, 2, 3] }}")
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		p := newParser(elements[0].(*tokenizer.Mustache).Tokens)
//		expr, err := p.parse()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		compiler := NewCompiler(expr.Expr)
//		chunk, err := compiler.Compile()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		dumpChunk(chunk)
//	})
//}

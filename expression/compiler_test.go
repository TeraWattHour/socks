package expression

//func TestCompiler(t *testing.T) {
//	t.Run("dot access", func(t *testing.T) {
//		file := helpers.File{
//			Name:    "debug.txt",
//			Content: "{{ a.b.c }}",
//		}
//
//		elements, err := Tokenize("debug.txt", file.Content)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		expr, err := Parse(file, elements[0].(*tokenizer.Mustache).Tokens)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		compiler := NewCompiler(file, expr.Expr)
//		chunk, err := compiler.Compile()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		dumpChunk(chunk)
//	})
//}

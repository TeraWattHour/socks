package expression

func Create(source string) (*VM, error) {
	tokens, err := Tokenize(source)
	if err != nil {
		return nil, err
	}

	ast, err := Parse(tokens)
	if err != nil {
		return nil, err
	}

	chunk, err := NewCompiler(ast.Expr).Compile()
	if err != nil {
		return nil, err
	}

	return NewVM(chunk), nil
}

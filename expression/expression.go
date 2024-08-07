package expression

func Create(source string) (*VM, []string, error) {
	tokens, err := Tokenize(source)
	if err != nil {
		return nil, nil, err
	}

	ast, err := Parse(tokens)
	if err != nil {
		return nil, nil, err
	}

	chunk, err := NewCompiler(ast.Expr).Compile()
	if err != nil {
		return nil, nil, err
	}

	return NewVM(chunk), ast.Dependencies, nil
}

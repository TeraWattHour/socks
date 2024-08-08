package expression

import "github.com/terawatthour/socks/internal/helpers"

func Create(source string, blockLocation helpers.Location) (*VM, []string, error) {
	tokens, err := Tokenize(source, blockLocation)
	if err != nil {
		return nil, nil, err
	}

	ast, err := Parse(tokens)
	if err != nil {
		return nil, nil, err
	}

	program, err := NewCompiler(ast.Expr).Compile()
	if err != nil {
		return nil, nil, err
	}

	return NewVM(program), ast.Dependencies, nil
}

package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    Expression
		wantErr bool
	}{
		{
			name: "chaining (just identifiers)",
			expr: "a.b.c",
			want: &VariableAccess{
				Left: &VariableAccess{
					Left:  &Identifier{Value: "a"},
					Right: &Identifier{Value: "b"},
				},
				Right: &Identifier{Value: "c"},
			},
			wantErr: false,
		},
		{
			name: "chaining (combined with array access)",
			expr: "a.b[1].c.d",
			want: &VariableAccess{
				Left: &VariableAccess{
					Left: &ArrayAccess{
						Accessed: &VariableAccess{
							Left:  &Identifier{Value: "a"},
							Right: &Identifier{Value: "b"},
						},
						Index: &Integer{Value: 1},
					},
					Right: &Identifier{Value: "c"},
				},
				Right: &Identifier{Value: "d"},
			},
			wantErr: false,
		},
		{
			name: "chaining (combined with method calls)",
			expr: "a.b().c[2]",
			want: &ArrayAccess{
				Accessed: &VariableAccess{
					Left: &FunctionCall{
						Called: &VariableAccess{
							Left:  &Identifier{Value: "a"},
							Right: &Identifier{Value: "b"},
						},
					},
					Right: &Identifier{Value: "c"},
				},
				Index: &Integer{Value: 2},
			},
		},
		{
			name: "recognize builtins",
			expr: "functionCall().int(int(123))",
			want: &FunctionCall{
				Called: &VariableAccess{
					Left: &FunctionCall{
						Called: &Identifier{Value: "functionCall"},
					},
					Right: &Identifier{Value: "int"},
				},
				Args: []Expression{
					&Builtin{
						Name: "int",
						Args: []Expression{
							&Integer{Value: 123},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		elements, err := tokenizer.Tokenize(fmt.Sprintf("{{ %s }}", tt.expr))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		p := newParser(elements[0].(*tokenizer.Mustache).Tokens)
		got, err := p.parser()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. parser() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !got.Expr.IsEqual(tt.want) {
			t.Errorf("%q, got:\n%s\nexpected:\n%s\n", tt.name, got, tt.want)
		}
	}
}

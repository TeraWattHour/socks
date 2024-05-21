package expression

import (
	"fmt"
	"github.com/terawatthour/socks/tokenizer"
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
			want: &Chain{
				Left: &Chain{
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
			want: &Chain{
				Left: &Chain{
					Left: &FieldAccess{
						Accessed: &Chain{
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
			want: &FieldAccess{
				Accessed: &Chain{
					Left: &FunctionCall{
						Called: &Chain{
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
				Called: &Chain{
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
		{
			name: "elvis operator",
			expr: "a ?: b * 2 ?: c + 1",
			want: &InfixExpression{
				Left: &InfixExpression{Left: &Identifier{Value: "a"}, Op: tokenizer.TokElvis, Right: &InfixExpression{
					Left:  &Identifier{Value: "b"},
					Op:    tokenizer.TokAsterisk,
					Right: &Integer{Value: 2},
				}},
				Op:    tokenizer.TokElvis,
				Right: &InfixExpression{Left: &Identifier{Value: "c"}, Op: tokenizer.TokPlus, Right: &Integer{Value: 1}},
			},
		},
		{
			"not precedence 1",
			"not (a in b)",
			&PrefixExpression{
				Op: tokenizer.TokNot,
				Right: &InfixExpression{
					Left:  &Identifier{Value: "a"},
					Op:    tokenizer.TokIn,
					Right: &Identifier{Value: "b"},
				},
			},
			false,
		},
		{
			"not precedence 2",
			"not a in b",
			&InfixExpression{
				Left: &PrefixExpression{
					Op:    tokenizer.TokNot,
					Right: &Identifier{Value: "a"},
				},
				Op:    tokenizer.TokIn,
				Right: &Identifier{Value: "b"},
			},
			false,
		},
		{
			"ternary",
			"a ? b ** 2 : c",
			&Ternary{
				Condition:   &Identifier{Value: "a"},
				Consequence: &InfixExpression{Left: &Identifier{Value: "b"}, Op: tokenizer.TokPower, Right: &Integer{Value: 2}},
				Alternative: &Identifier{Value: "c"},
			},
			false,
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

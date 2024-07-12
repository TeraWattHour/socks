package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want Expression
	}{
		{
			name: "arithmetics",
			expr: "1 + 2 * 3 ** 3 * (2 + 4)",
			want: &InfixExpression{
				Left: &Integer{Value: 1},
				Op:   tokenizer.TokPlus,
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left: &Integer{Value: 2},
						Op:   tokenizer.TokAsterisk,
						Right: &InfixExpression{
							Left:  &Integer{Value: 3},
							Op:    tokenizer.TokPower,
							Right: &Integer{Value: 3},
						},
					},
					Op: tokenizer.TokAsterisk,
					Right: &InfixExpression{
						Left:  &Integer{Value: 2},
						Op:    tokenizer.TokPlus,
						Right: &Integer{Value: 4},
					},
				},
			},
		},
		{
			name: "chaining (just identifiers)",
			expr: "a.b(1, 2, 3+4).c",
			want: &Chain{
				Parts: []Expression{
					&Identifier{Value: "a"},
					&DotAccess{Property: "b"},
					&FunctionCall{Args: []Expression{
						&Integer{Value: 1},
						&Integer{Value: 2},
						&InfixExpression{Left: &Integer{Value: 3}, Op: tokenizer.TokPlus, Right: &Integer{Value: 4}},
					}},
					&DotAccess{Property: "c"},
				},
			},
		},
		{
			name: "chaining (combined with array access)",
			expr: "a.b[1].c.d",
			want: &Chain{
				Parts: []Expression{
					&Identifier{Value: "a"},
					&DotAccess{Property: "b"},
					&FieldAccess{Index: &Integer{Value: 1}},
					&DotAccess{Property: "c"},
					&DotAccess{Property: "d"},
				},
			},
		},
		{
			name: "chaining (combined with method calls)",
			expr: "a.b(1).c[function(23.2, abc())].d",
			want: &Chain{
				Parts: []Expression{
					&Identifier{Value: "a"},
					&DotAccess{Property: "b"},
					&FunctionCall{Args: []Expression{&Integer{Value: 1}}},
					&DotAccess{Property: "c"},
					&FieldAccess{Index: &Chain{
						Parts: []Expression{
							&Identifier{Value: "function"},
							&FunctionCall{Args: []Expression{&Float{Value: 23.2}, &Chain{Parts: []Expression{
								&Identifier{Value: "abc"},
								&FunctionCall{},
							}}},
							},
						}},
					},
					&DotAccess{Property: "d"},
				},
			},
		},
		{
			name: "recognize builtins",
			expr: "int(int(123))",
			want: &Chain{
				Parts: []Expression{
					&Identifier{Value: "int"},
					&FunctionCall{Args: []Expression{
						&Chain{Parts: []Expression{&Identifier{Value: "int"}, &FunctionCall{Args: []Expression{&Integer{Value: 123}}}}},
					}},
				},
			},
		},
		{
			name: "array access",
			expr: "a[1]?.(a[a], b(), \"test\"[1])",
			want: &Chain{
				Parts: []Expression{
					&Identifier{Value: "a"},
					&FieldAccess{Index: &Integer{Value: 1}},
					&OptionalAccess{},
					&FunctionCall{
						Args: []Expression{
							&Chain{
								Parts: []Expression{
									&Identifier{Value: "a"},
									&FieldAccess{Index: &Identifier{Value: "a"}},
								},
							},
							&Chain{Parts: []Expression{&Identifier{Value: "b"}, &FunctionCall{}}},
							&Chain{Parts: []Expression{
								&StringLiteral{Value: "test"},
								&FieldAccess{Index: &Integer{Value: 1}},
							}},
						},
					},
				},
			},
		},
		{
			name: "elvis operator",
			expr: "a ?: b * 2 ?: c + 1",
			want: &InfixExpression{
				Left: &Identifier{Value: "a"},
				Op:   tokenizer.TokElvis,
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left:  &Identifier{Value: "b"},
						Op:    tokenizer.TokAsterisk,
						Right: &Integer{Value: 2},
					},
					Op: tokenizer.TokElvis,
					Right: &InfixExpression{
						Left:  &Identifier{Value: "c"},
						Op:    tokenizer.TokPlus,
						Right: &Integer{Value: 1},
					},
				},
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
		},
		{
			"array literal",
			"[1, 2, 3][1]",
			&Chain{
				Parts: []Expression{
					&Array{Items: []Expression{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}}},
					&FieldAccess{Index: &Integer{Value: 1}},
				}},
		},
		{
			"ternary",
			"a ? b ** 2 : c",
			&Ternary{
				Condition:   &Identifier{Value: "a"},
				Consequence: &InfixExpression{Left: &Identifier{Value: "b"}, Op: tokenizer.TokPower, Right: &Integer{Value: 2}},
				Alternative: &Identifier{Value: "c"},
			},
		},
		{
			"fallback chain",
			"(a ?: b)?.c",
			&Chain{
				Parts: []Expression{
					&InfixExpression{
						Left:  &Identifier{Value: "a"},
						Op:    tokenizer.TokElvis,
						Right: &Identifier{Value: "b"},
					},
					&OptionalAccess{},
					&Identifier{Value: "c"},
				},
			},
		},
		{
			"elvis I",
			"voidMember?.method() ?: 1 != 1 ?: 420",
			&InfixExpression{
				Left: &Chain{
					Parts: []Expression{
						&Identifier{Value: "voidMember"},
						&OptionalAccess{},
						&Identifier{Value: "method"},
						&FunctionCall{Args: []Expression{}},
					},
				},
				Op: tokenizer.TokElvis,
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left:  &Integer{Value: 1},
						Op:    tokenizer.TokNeq,
						Right: &Integer{Value: 1},
					},
					Op:    tokenizer.TokElvis,
					Right: &Integer{Value: 420},
				},
			},
		},
		{
			"elvis operator",
			"false ? a ?: b == b : 123",
			&Ternary{
				Condition: &Boolean{Value: false},
				Consequence: &InfixExpression{
					Left: &Identifier{Value: "a"},
					Op:   tokenizer.TokElvis,
					Right: &InfixExpression{
						Left:  &Identifier{Value: "b"},
						Op:    tokenizer.TokEq,
						Right: &Identifier{Value: "b"},
					},
				},
				Alternative: &Integer{Value: 123},
			},
		},
		{
			"elvis and chaining",
			"a ?: b.c",
			&InfixExpression{
				Left: &Identifier{Value: "a"},
				Op:   tokenizer.TokElvis,
				Right: &Chain{
					Parts: []Expression{
						&Identifier{Value: "b"},
						&DotAccess{Property: "c"},
					},
				},
			},
		},
		{
			"nil comparison",
			"a == nil ? false : true",
			&Ternary{
				Condition: &InfixExpression{
					Left:  &Identifier{Value: "a"},
					Op:    tokenizer.TokEq,
					Right: &Nil{},
				},
				Consequence: &Boolean{Value: false},
				Alternative: &Boolean{Value: true},
			},
		},
	}
	for _, tt := range tests {
		elements, err := tokenizer.Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", tt.expr))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		got, err := Parse(helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", tt.expr)}, elements[0].(*tokenizer.Mustache).Tokens)
		if err != nil {
			t.Errorf("unexpected error:\n%s", err)
			continue
		}

		if !got.Expr.IsEqual(tt.want) {
			t.Errorf("%q, got:\n%s\nexpected:\n%s\n", tt.name, got.Expr.Literal(), tt.want.Literal())
		}
	}
}

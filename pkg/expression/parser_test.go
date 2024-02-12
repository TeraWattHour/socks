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
			name: "simple expression",
			expr: "true and false",
			want: &InfixExpression{
				Left:  &Boolean{Value: true},
				Op:    "and",
				Right: &Boolean{Value: false},
			},
			wantErr: false,
		},
		{
			name: "simple expression",
			expr: "true or false",
			want: &InfixExpression{
				Left:  &Boolean{Value: true},
				Op:    "or",
				Right: &Boolean{Value: false},
			},
			wantErr: false,
		},
		{
			name: "simple expression",
			expr: "true or false and true",
			want: &InfixExpression{
				Left: &Boolean{Value: true},
				Op:   "or",
				Right: &InfixExpression{
					Left:  &Boolean{Value: false},
					Op:    "and",
					Right: &Boolean{Value: true},
				},
			},
			wantErr: false,
		},
		{
			name: "simple expression",
			expr: "true and false or 7 // 2 == 1",
			want: &InfixExpression{
				Left: &InfixExpression{
					Left:  &Boolean{Value: true},
					Op:    "and",
					Right: &Boolean{Value: false},
				},
				Op: "or",
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left:  &Numeric{Value: 7},
						Op:    "//",
						Right: &Numeric{Value: 2},
					},
					Op:    "==",
					Right: &Numeric{Value: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "simple expression",
			expr: "true and false or true and false",
			want: &InfixExpression{
				Left: &InfixExpression{
					Left:  &Boolean{Value: true},
					Op:    "and",
					Right: &Boolean{Value: false},
				},
				Op: "or",
				Right: &InfixExpression{
					Left:  &Boolean{Value: true},
					Op:    "and",
					Right: &Boolean{Value: false},
				},
			},
			wantErr: false,
		},
		{
			"string literals",
			`"hello" + "world"`,
			&InfixExpression{
				Left:  &StringLiteral{Value: "hello"},
				Op:    "+",
				Right: &StringLiteral{Value: "world"},
			},
			false,
		},
		{
			name: "algebraic expression",
			expr: "1 ** 123 / 2",
			want: &InfixExpression{
				Left: &InfixExpression{
					Left:  &Numeric{Value: 1},
					Op:    "**",
					Right: &Numeric{Value: 123},
				},
				Op:    "/",
				Right: &Numeric{Value: 2},
			},
		},
		{
			name: "algebraic expression with idents",
			expr: "1 + 123 / 2 + 1.23 * constant",
			want: &InfixExpression{
				Left: &InfixExpression{
					Left: &Numeric{Value: 1},
					Op:   "+",
					Right: &InfixExpression{
						Left: &Numeric{Value: 123},
						Op:   "/",
						Right: &Numeric{
							Value: 2,
						},
					},
				},
				Op: "+",
				Right: &InfixExpression{
					Left: &Numeric{Value: 1.23},
					Op:   "*",
					Right: &Identifier{
						Value: "constant",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "negation",
			expr: "not(1 not in [2] and 1 in [2])",
			want: &PrefixExpression{
				Op: "not",
				Right: &InfixExpression{
					Left: &PrefixExpression{
						Op: "not",
						Right: &InfixExpression{
							Left: &Numeric{Value: 1},
							Op:   "in",
							Right: &Array{Items: []Expression{
								&Numeric{Value: 2},
							}},
						},
					},
					Op: "and",
					Right: &InfixExpression{
						Left: &Numeric{Value: 1},
						Op:   "in",
						Right: &Array{
							Items: []Expression{
								&Numeric{Value: 2},
							}},
					},
				},
			},
		},
		{
			"negation",
			"not(1 not in [2] and 1 in [2])",
			&PrefixExpression{
				Op: "not",
				Right: &InfixExpression{
					Left: &PrefixExpression{
						Op: "not",
						Right: &InfixExpression{
							Left:  &Numeric{Value: 1},
							Op:    "in",
							Right: &Array{Items: []Expression{&Numeric{Value: 2}}},
						},
					},
					Op: "and",
					Right: &InfixExpression{
						Left:  &Numeric{Value: 1},
						Op:    "in",
						Right: &Array{Items: []Expression{&Numeric{Value: 2}}},
					},
				},
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

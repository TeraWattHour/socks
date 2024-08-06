package expression

//func TestParse(t *testing.T) {
//	tests := []struct {
//		name string
//		expr string
//		want Expression
//	}{
//		{
//			name: "arithmetics",
//			expr: "1 + 2 * 3 ** 3 * (2 + 4)",
//			want: &InfixExpression{
//				Left: &Integer{Value: 1},
//				Op:   TokPlus,
//				Right: &InfixExpression{
//					Left: &InfixExpression{
//						Left: &Integer{Value: 2},
//						Op:   TokAsterisk,
//						Right: &InfixExpression{
//							Left:  &Integer{Value: 3},
//							Op:    TokPower,
//							Right: &Integer{Value: 3},
//						},
//					},
//					Op: TokAsterisk,
//					Right: &InfixExpression{
//						Left:  &Integer{Value: 2},
//						Op:    TokPlus,
//						Right: &Integer{Value: 4},
//					},
//				},
//			},
//		},
//		{
//			name: "chaining (just identifiers)",
//			expr: "a.b(1, 2, 3+4).c",
//			want: &Chain{
//				Parts: []Expression{
//					&Identifier{Value: "a"},
//					&DotAccess{Property: "b"},
//					&FunctionCall{Args: []Expression{
//						&Integer{Value: 1},
//						&Integer{Value: 2},
//						&InfixExpression{Left: &Integer{Value: 3}, Op: TokPlus, Right: &Integer{Value: 4}},
//					}},
//					&DotAccess{Property: "c"},
//				},
//			},
//		},
//		{
//			name: "chaining (combined with array access)",
//			expr: "a.b[1].c.d",
//			want: &Chain{
//				Parts: []Expression{
//					&Identifier{Value: "a"},
//					&DotAccess{Property: "b"},
//					&FieldAccess{Index: &Integer{Value: 1}},
//					&DotAccess{Property: "c"},
//					&DotAccess{Property: "d"},
//				},
//			},
//		},
//		{
//			name: "chaining (combined with method calls)",
//			expr: "a.b(1).c[function(23.2, abc())].d",
//			want: &Chain{
//				Parts: []Expression{
//					&Identifier{Value: "a"},
//					&DotAccess{Property: "b"},
//					&FunctionCall{Args: []Expression{&Integer{Value: 1}}},
//					&DotAccess{Property: "c"},
//					&FieldAccess{Index: &Chain{
//						Parts: []Expression{
//							&Identifier{Value: "function"},
//							&FunctionCall{Args: []Expression{&Float{Value: 23.2}, &Chain{Parts: []Expression{
//								&Identifier{Value: "abc"},
//								&FunctionCall{},
//							}}},
//							},
//						}},
//					},
//					&DotAccess{Property: "d"},
//				},
//			},
//		},
//		{
//			name: "recognize builtins",
//			expr: "int(int(123))",
//			want: &Chain{
//				Parts: []Expression{
//					&Identifier{Value: "int"},
//					&FunctionCall{Args: []Expression{
//						&Chain{Parts: []Expression{&Identifier{Value: "int"}, &FunctionCall{Args: []Expression{&Integer{Value: 123}}}}},
//					}},
//				},
//			},
//		},
//		{
//			name: "array access",
//			expr: "a[1]?.(a[a], b(), \"test\"[1])",
//			want: &Chain{
//				Parts: []Expression{
//					&Identifier{Value: "a"},
//					&FieldAccess{Index: &Integer{Value: 1}},
//					&OptionalAccess{},
//					&FunctionCall{
//						Args: []Expression{
//							&Chain{
//								Parts: []Expression{
//									&Identifier{Value: "a"},
//									&FieldAccess{Index: &Identifier{Value: "a"}},
//								},
//							},
//							&Chain{Parts: []Expression{&Identifier{Value: "b"}, &FunctionCall{}}},
//							&Chain{Parts: []Expression{
//								&StringLiteral{Value: "test"},
//								&FieldAccess{Index: &Integer{Value: 1}},
//							}},
//						},
//					},
//				},
//			},
//		},
//		{
//			name: "elvis operator",
//			expr: "a ?: b * 2 ?: c + 1",
//			want: &InfixExpression{
//				Left: &Identifier{Value: "a"},
//				Op:   TokElvis,
//				Right: &InfixExpression{
//					Left: &InfixExpression{
//						Left:  &Identifier{Value: "b"},
//						Op:    TokAsterisk,
//						Right: &Integer{Value: 2},
//					},
//					Op: TokElvis,
//					Right: &InfixExpression{
//						Left:  &Identifier{Value: "c"},
//						Op:    TokPlus,
//						Right: &Integer{Value: 1},
//					},
//				},
//			},
//		},
//		{
//			"not precedence 1",
//			"not (a in b)",
//			&PrefixExpression{
//				Op: TokNot,
//				Right: &InfixExpression{
//					Left:  &Identifier{Value: "a"},
//					Op:    TokIn,
//					Right: &Identifier{Value: "b"},
//				},
//			},
//		},
//		{
//			"not precedence 2",
//			"not a in b",
//			&InfixExpression{
//				Left: &PrefixExpression{
//					Op:    TokNot,
//					Right: &Identifier{Value: "a"},
//				},
//				Op:    TokIn,
//				Right: &Identifier{Value: "b"},
//			},
//		},
//		{
//			"array literal",
//			"[1, 2, 3][1]",
//			&Chain{
//				Parts: []Expression{
//					&Array{Items: []Expression{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}}},
//					&FieldAccess{Index: &Integer{Value: 1}},
//				}},
//		},
//		{
//			"ternary",
//			"a ? b ** 2 : c",
//			&Ternary{
//				Condition:   &Identifier{Value: "a"},
//				Consequence: &InfixExpression{Left: &Identifier{Value: "b"}, Op: TokPower, Right: &Integer{Value: 2}},
//				Alternative: &Identifier{Value: "c"},
//			},
//		},
//		{
//			"fallback chain",
//			"(a ?: b)?.c",
//			&Chain{
//				Parts: []Expression{
//					&InfixExpression{
//						Left:  &Identifier{Value: "a"},
//						Op:    TokElvis,
//						Right: &Identifier{Value: "b"},
//					},
//					&OptionalAccess{},
//					&Identifier{Value: "c"},
//				},
//			},
//		},
//		{
//			"elvis I",
//			"voidMember?.method() ?: 1 != 1 ?: 420",
//			&InfixExpression{
//				Left: &Chain{
//					Parts: []Expression{
//						&Identifier{Value: "voidMember"},
//						&OptionalAccess{},
//						&Identifier{Value: "method"},
//						&FunctionCall{Args: []Expression{}},
//					},
//				},
//				Op: TokElvis,
//				Right: &InfixExpression{
//					Left: &InfixExpression{
//						Left:  &Integer{Value: 1},
//						Op:    TokNeq,
//						Right: &Integer{Value: 1},
//					},
//					Op:    TokElvis,
//					Right: &Integer{Value: 420},
//				},
//			},
//		},
//		{
//			"elvis operator",
//			"false ? a ?: b == b : 123",
//			&Ternary{
//				Condition: &Boolean{Value: false},
//				Consequence: &InfixExpression{
//					Left: &Identifier{Value: "a"},
//					Op:   TokElvis,
//					Right: &InfixExpression{
//						Left:  &Identifier{Value: "b"},
//						Op:    TokEq,
//						Right: &Identifier{Value: "b"},
//					},
//				},
//				Alternative: &Integer{Value: 123},
//			},
//		},
//		{
//			"elvis and chaining",
//			"a ?: b.c",
//			&InfixExpression{
//				Left: &Identifier{Value: "a"},
//				Op:   TokElvis,
//				Right: &Chain{
//					Parts: []Expression{
//						&Identifier{Value: "b"},
//						&DotAccess{Property: "c"},
//					},
//				},
//			},
//		},
//		{
//			"nil comparison",
//			"a == nil ? false : true",
//			&Ternary{
//				Condition: &InfixExpression{
//					Left:  &Identifier{Value: "a"},
//					Op:    TokEq,
//					Right: &Nil{},
//				},
//				Consequence: &Boolean{Value: false},
//				Alternative: &Boolean{Value: true},
//			},
//		},
//	}
//	for _, tt := range tests {
//		elements, err := Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", tt.expr))
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		got, err := Parse(helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", tt.expr)}, elements[0].(*tokenizer.Mustache).Tokens)
//		if err != nil {
//			t.Errorf("unexpected error:\n%s", err)
//			continue
//		}
//
//		if !got.Expr.IsEqual(tt.want) {
//			t.Errorf("%q, got:\n%s\nexpected:\n%s\n", tt.name, got.Expr.Literal(), tt.want.Literal())
//		}
//	}
//}

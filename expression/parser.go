package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type parser struct {
	file          helpers.File
	previousToken tokenizer.Token
	currentToken  tokenizer.Token
	nextToken     tokenizer.Token
	tokens        []tokenizer.Token
	cursor        int
	dependencies  []string
}

func Parse(file helpers.File, tokens []tokenizer.Token) (*WrappedExpression, error) {
	parser := &parser{
		file:         file,
		cursor:       -1,
		tokens:       tokens,
		dependencies: make([]string, 0),
	}
	return parser.parse()
}

func (p *parser) parse() (*WrappedExpression, error) {
	p.advance()

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if p.nextToken.Kind != tokenizer.TokEmpty {
		return nil, p.error(
			fmt.Sprintf("unexpected %s", p.nextToken.Kind),
			p.nextToken.Location,
		)
	}

	return &WrappedExpression{
		Expr:         expr,
		Dependencies: p.dependencies,
	}, nil
}

func (p *parser) expression() (Expression, error) {
	return p.ternary()
}

func (p *parser) ternary() (Expression, error) {
	left, err := p.elvis()
	if err != nil {
		return nil, err
	}

	token := p.currentToken

	if p.currentIs(tokenizer.TokQuestion) {
		p.advance()
		trueExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if !p.currentIs(tokenizer.TokColon) {
			return nil, p.error(fmt.Sprintf("unexpected %s", p.currentToken.Kind), p.currentToken.Location)
		}

		p.advance()
		falseExpr, err := p.ternary()
		if err != nil {
			return nil, err
		}

		left = &Ternary{
			Token:       token,
			Condition:   left,
			Consequence: trueExpr,
			Alternative: falseExpr,
		}
	}

	return left, nil
}

func (p *parser) elvis() (Expression, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	token := p.currentToken

	if p.currentIs(tokenizer.TokElvis) {
		p.advance()
		defaultExpr, err := p.elvis()
		if err != nil {
			return nil, err
		}

		expr = &InfixExpression{
			Token: token,
			Op:    token.Kind,
			Left:  expr,
			Right: defaultExpr,
		}
	}

	return expr, nil
}

func (p *parser) comparison() (Expression, error) {
	left, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.currentIs(tokenizer.TokLt, tokenizer.TokGt, tokenizer.TokLte, tokenizer.TokGte, tokenizer.TokIn) {
		op := p.advance()
		right, err := p.term()
		if err != nil {
			return nil, err
		}

		left = &InfixExpression{
			Token: op,
			Op:    op.Kind,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *parser) equality() (Expression, error) {
	left, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.currentIs(tokenizer.TokEq, tokenizer.TokNeq) {
		op := p.advance()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		left = &InfixExpression{
			Token: op,
			Op:    op.Kind,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *parser) term() (Expression, error) {
	left, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.currentIs(tokenizer.TokPlus, tokenizer.TokMinus) {
		op := p.advance()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		left = &InfixExpression{
			Token: op,
			Op:    op.Kind,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *parser) factor() (Expression, error) {
	left, err := p.power()
	if err != nil {
		return nil, err
	}

	for p.currentIs(tokenizer.TokAsterisk, tokenizer.TokSlash, tokenizer.TokModulo) {
		op := p.advance()
		right, err := p.power()
		if err != nil {
			return nil, err
		}

		left = &InfixExpression{
			Token: op,
			Op:    op.Kind,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *parser) power() (Expression, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.currentIs(tokenizer.TokPower) {
		op := p.advance()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		left = &InfixExpression{
			Token: op,
			Op:    op.Kind,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *parser) unary() (Expression, error) {
	if p.currentIs(tokenizer.TokMinus, tokenizer.TokBang, tokenizer.TokNot) {
		op := p.advance()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &PrefixExpression{
			Token: op,
			Op:    op.Kind,
			Right: right,
		}, nil
	}

	return p.primary()
}

func (p *parser) primary() (Expression, error) {
	token := p.advance()

	switch token.Kind {
	case tokenizer.TokFalse:
		return &Boolean{Token: token, Value: false}, nil
	case tokenizer.TokTrue:
		return &Boolean{Token: token, Value: true}, nil
	case tokenizer.TokNil:
		return &Nil{Token: token}, nil
	case tokenizer.TokNumeric:
		return p.numeric(token)
	case tokenizer.TokString:
		return p.string(token)
	case tokenizer.TokIdent:
		return p.identifier(token)
	case tokenizer.TokLbrack:
		return p.array()
	case tokenizer.TokLparen:
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if !p.currentIs(tokenizer.TokRparen) {
			return nil, p.error("expected `)`", p.currentToken.Location)
		}

		p.advance()

		if p.currentIs(tokenizer.TokLbrack, tokenizer.TokLparen, tokenizer.TokDot, tokenizer.TokOptionalChain) {
			return p.chain(expr)
		}

		return expr, nil
	default:
		return nil, p.error("unexpected token "+p.currentToken.Literal, p.currentToken.Location)
	}
}

func (p *parser) chain(left Expression) (Expression, error) {
	expr := &Chain{
		Token: p.currentToken,
		Parts: []Expression{left},
	}

	for p.currentIs(tokenizer.TokDot, tokenizer.TokOptionalChain, tokenizer.TokLparen, tokenizer.TokLbrack) {
		token := p.currentToken
		switch token.Kind {
		case tokenizer.TokDot:
			if !p.expectNext(tokenizer.TokIdent) {
				return nil, p.error(fmt.Sprintf("unexpected %s, expected identifier", p.nextToken.Kind), p.nextToken.Location)
			}

			expr.Parts.Push(&DotAccess{Token: p.currentToken, Property: p.currentToken.Literal})
		case tokenizer.TokOptionalChain:
			expr.Parts.Push(&OptionalAccess{Token: p.currentToken})

			if !p.nextIs(tokenizer.TokIdent, tokenizer.TokLparen, tokenizer.TokLbrack) {
				return nil, p.error(fmt.Sprintf("unexpected %s, expected identifier, method call, or array access", p.nextToken.Kind), p.nextToken.Location)
			}

			if p.expectNext(tokenizer.TokIdent) {
				expr.Parts.Push(&Identifier{Token: p.currentToken, Value: p.currentToken.Literal})
			}
		case tokenizer.TokLparen:
			p.advance()

			args, err := p.list(tokenizer.TokRparen, ")")
			if err != nil {
				return nil, err
			}
			assert(p.currentToken.Kind == tokenizer.TokRparen, "p.currentToken after p.list must always be the end literal")

			expr.Parts.Push(&FunctionCall{Token: token, Args: args, closeToken: p.currentToken})
		case tokenizer.TokLbrack:
			p.advance()

			index, err := p.expression()
			if err != nil {
				return nil, err
			}

			if !p.currentIs(tokenizer.TokRbrack) {
				return nil, p.error("expected `]`", p.currentToken.Location)
			}

			expr.Parts.Push(&FieldAccess{Token: token, Index: index, closeToken: p.currentToken})
		default:
			panic("unreachable")
		}

		p.advance()
	}

	return expr, nil
}

func (p *parser) array() (Expression, error) {
	var err error
	array := &Array{Token: p.currentToken}
	array.Items, err = p.list(tokenizer.TokRbrack, "]")
	if err != nil {
		return nil, err
	}
	assert(p.currentToken.Kind == tokenizer.TokRbrack, "p.currentToken after p.list must always be the end literal")

	if p.nextIs(tokenizer.TokLbrack) {
		p.advance()
		return p.chain(array)
	}

	return array, err
}

func (p *parser) list(end tokenizer.TokenKind, endLiteral string) ([]Expression, error) {
	list := make([]Expression, 0)
	if p.currentIs(end) {
		return list, nil
	}

	listElement, err := p.expression()
	if err != nil {
		return nil, err
	}
	list = append(list, listElement)

	for p.currentIs(tokenizer.TokComma) {
		p.advance()
		listElement, err = p.expression()
		if err != nil {
			return nil, err
		}
		list = append(list, listElement)
	}

	if !p.currentIs(end) {
		return nil, p.error(fmt.Sprintf("expected `%s`", endLiteral), p.currentToken.Location)
	}

	return list, nil
}

func (p *parser) identifier(token tokenizer.Token) (Expression, error) {
	if !slices.ContainsFunc(builtinNames, func(value reflect.Value) bool {
		return value.String() == token.Literal
	}) {
		p.dependencies = append(p.dependencies, token.Literal)
	}

	ident := &Identifier{Token: token, Value: token.Literal}

	if p.currentIs(tokenizer.TokLparen, tokenizer.TokLbrack, tokenizer.TokDot, tokenizer.TokOptionalChain) {
		return p.chain(ident)
	}

	return ident, nil
}

func (p *parser) string(token tokenizer.Token) (Expression, error) {
	str := &StringLiteral{Token: token, Value: token.Literal}

	if p.currentIs(tokenizer.TokLbrack) {
		return p.chain(str)
	}

	return str, nil
}

func (p *parser) numeric(token tokenizer.Token) (Expression, error) {
	if strings.Contains(token.Literal, ".") {
		f64, err := strconv.ParseFloat(token.Literal, 64)
		if err != nil {
			return nil, p.error("malformed float literal", token.Location)
		}
		return &Float{Token: token, Value: f64}, nil
	}

	i64, err := strconv.ParseInt(token.Literal, 0, 64)
	if err != nil {
		return nil, p.error("malformed integer literal", token.Location)
	}

	return &Integer{Token: token, Value: int(i64)}, nil
}

func (p *parser) nextIs(kinds ...tokenizer.TokenKind) bool {
	for _, kind := range kinds {
		if kind == p.nextToken.Kind {
			return true
		}
	}
	return false
}

func (p *parser) currentIs(kinds ...tokenizer.TokenKind) bool {
	for _, kind := range kinds {
		if kind == p.currentToken.Kind {
			return true
		}
	}
	return false
}

func (p *parser) expectNext(kind tokenizer.TokenKind) bool {
	if p.nextToken.Kind == kind {
		p.advance()
		return true
	}
	return false
}

func (p *parser) advance() tokenizer.Token {
	p.previousToken = p.currentToken

	p.cursor++
	if p.cursor >= len(p.tokens) {
		p.currentToken = tokenizer.Token{}
		p.nextToken = tokenizer.Token{}
		return p.previousToken
	}

	p.currentToken = p.tokens[p.cursor]

	if p.cursor+1 >= len(p.tokens) {
		p.nextToken = tokenizer.Token{}
	} else {
		p.nextToken = p.tokens[p.cursor+1]
	}

	return p.previousToken
}

func (p *parser) error(message string, location helpers.Location) error {
	return errors2.New(message, p.file.Name, p.file.Content, location, location.FromOther())
}

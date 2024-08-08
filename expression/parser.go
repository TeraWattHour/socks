package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type parser struct {
	previousToken Token
	currentToken  Token
	nextToken     Token
	tokens        []Token
	cursor        int
	dependencies  []string
}

func Parse(tokens []Token) (*WrappedExpression, error) {
	parser := &parser{
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

	if p.nextToken.Kind != TokEmpty {
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

	if p.currentIs(TokQuestion) {
		p.advance()
		trueExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if !p.currentIs(TokColon) {
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

	if p.currentIs(TokElvis) {
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

	for p.currentIs(TokLt, TokGt, TokLte, TokGte, TokIn) {
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

	for p.currentIs(TokEq, TokNeq) {
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

	for p.currentIs(TokPlus, TokMinus) {
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

	for p.currentIs(TokAsterisk, TokSlash, TokModulo) {
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

	for p.currentIs(TokPower) {
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
	if p.currentIs(TokMinus, TokBang, TokNot) {
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
	case TokFalse:
		return &Boolean{Token: token, Value: false}, nil
	case TokTrue:
		return &Boolean{Token: token, Value: true}, nil
	case TokNil:
		return &Nil{Token: token}, nil
	case TokNumeric:
		return p.numeric(token)
	case TokString:
		return p.string(token)
	case TokIdent:
		return p.identifier(token)
	case TokLbrack:
		return p.array()
	case TokLparen:
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		if !p.currentIs(TokRparen) {
			return nil, p.error("expected `)`", p.currentToken.Location)
		}

		p.advance()

		if p.currentIs(TokLbrack, TokLparen, TokDot, TokOptionalChain) {
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

	for p.currentIs(TokDot, TokOptionalChain, TokLparen, TokLbrack) {
		token := p.currentToken
		switch token.Kind {
		case TokDot:
			if !p.expectNext(TokIdent) {
				return nil, p.error(fmt.Sprintf("unexpected %s, expected identifier", p.nextToken.Kind), p.nextToken.Location)
			}

			expr.Parts.Push(&DotAccess{Token: p.currentToken, Property: p.currentToken.Literal})
		case TokOptionalChain:
			expr.Parts.Push(&OptionalAccess{Token: p.currentToken})

			if !p.nextIs(TokIdent, TokLparen, TokLbrack) {
				return nil, p.error(fmt.Sprintf("unexpected %s, expected identifier, method call, or array access", p.nextToken.Kind), p.nextToken.Location)
			}

			if p.expectNext(TokIdent) {
				expr.Parts.Push(&Identifier{Token: p.currentToken, Value: p.currentToken.Literal})
			}
		case TokLparen:
			p.advance()

			args, err := p.list(TokRparen, ")")
			if err != nil {
				return nil, err
			}
			assert(p.currentToken.Kind == TokRparen, "p.currentToken after p.list must always be the end literal")

			expr.Parts.Push(&FunctionCall{Token: token, Args: args, closeToken: p.currentToken})
		case TokLbrack:
			p.advance()

			index, err := p.expression()
			if err != nil {
				return nil, err
			}

			if !p.currentIs(TokRbrack) {
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
	array.Items, err = p.list(TokRbrack, "]")
	if err != nil {
		return nil, err
	}
	assert(p.currentToken.Kind == TokRbrack, "p.currentToken after p.list must always be the end literal")

	if p.nextIs(TokLbrack) {
		p.advance()
		return p.chain(array)
	}

	return array, err
}

func (p *parser) list(end TokenKind, endLiteral string) ([]Expression, error) {
	list := make([]Expression, 0)
	if p.currentIs(end) {
		return list, nil
	}

	listElement, err := p.expression()
	if err != nil {
		return nil, err
	}
	list = append(list, listElement)

	for p.currentIs(TokComma) {
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

func (p *parser) identifier(token Token) (Expression, error) {
	if !slices.ContainsFunc(builtinNames, func(value reflect.Value) bool {
		return value.String() == token.Literal
	}) {
		p.dependencies = append(p.dependencies, token.Literal)
	}

	ident := &Identifier{Token: token, Value: token.Literal}

	if p.currentIs(TokLparen, TokLbrack, TokDot, TokOptionalChain) {
		return p.chain(ident)
	}

	return ident, nil
}

func (p *parser) string(token Token) (Expression, error) {
	str := &StringLiteral{Token: token, Value: token.Literal}

	if p.currentIs(TokLbrack) {
		return p.chain(str)
	}

	return str, nil
}

func (p *parser) numeric(token Token) (Expression, error) {
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

func (p *parser) nextIs(kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if kind == p.nextToken.Kind {
			return true
		}
	}
	return false
}

func (p *parser) currentIs(kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if kind == p.currentToken.Kind {
			return true
		}
	}
	return false
}

func (p *parser) expectNext(kind TokenKind) bool {
	if p.nextToken.Kind == kind {
		p.advance()
		return true
	}
	return false
}

func (p *parser) advance() Token {
	p.previousToken = p.currentToken

	p.cursor++
	if p.cursor >= len(p.tokens) {
		p.currentToken = Token{}
		p.nextToken = Token{}
		return p.previousToken
	}

	p.currentToken = p.tokens[p.cursor]

	if p.cursor+1 >= len(p.tokens) {
		p.nextToken = Token{}
	} else {
		p.nextToken = p.tokens[p.cursor+1]
	}

	return p.previousToken
}

func (p *parser) error(message string, location helpers.Location) error {
	return errors2.New(message, location)
}

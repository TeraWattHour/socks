package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
)

type parser struct {
	file           helpers.File
	previousToken  *tokenizer.Token
	currentToken   *tokenizer.Token
	nextToken      *tokenizer.Token
	tokens         []tokenizer.Token
	prefixParseFns map[string]func() (Expression, error)
	infixParseFns  map[string]func(Expression) (Expression, error)
	cursor         int
	dependencies   []string
}

type Precedence int

const (
	_ Precedence = iota
	PrecLowest
	PrecOr
	PrecAnd
	PrecElvis
	PrecEqual
	PrecRelational
	PrecAdditive
	PrecMultiplicative
	PrecPower
	PrecPrefix
	PrecCall
	PrecChain
)

var precedences = map[string]Precedence{
	tokenizer.TokIdent: PrecLowest,

	tokenizer.TokOr: PrecOr,

	tokenizer.TokAnd: PrecAnd,

	tokenizer.TokElvis:    PrecElvis,
	tokenizer.TokQuestion: PrecElvis,

	tokenizer.TokEq:  PrecEqual,
	tokenizer.TokNeq: PrecEqual,

	tokenizer.TokLt:  PrecRelational,
	tokenizer.TokLte: PrecRelational,
	tokenizer.TokGt:  PrecRelational,
	tokenizer.TokGte: PrecRelational,
	tokenizer.TokIn:  PrecRelational,

	tokenizer.TokNot:   PrecAdditive,
	tokenizer.TokPlus:  PrecAdditive,
	tokenizer.TokMinus: PrecAdditive,

	tokenizer.TokAsterisk: PrecMultiplicative,
	tokenizer.TokSlash:    PrecMultiplicative,
	tokenizer.TokModulo:   PrecMultiplicative,

	tokenizer.TokPower: PrecPower,

	tokenizer.TokBang: PrecPrefix,

	tokenizer.TokDot:           PrecChain,
	tokenizer.TokOptionalChain: PrecChain,

	tokenizer.TokLparen: PrecCall,
	tokenizer.TokLbrack: PrecCall,
}

func Parse(file helpers.File, tokens []tokenizer.Token) (*WrappedExpression, error) {
	return newParser(file, tokens).parse()
}

func newParser(file helpers.File, tokens []tokenizer.Token) *parser {
	p := &parser{
		file:           file,
		cursor:         -1,
		tokens:         tokens,
		prefixParseFns: make(map[string]func() (Expression, error)),
		infixParseFns:  make(map[string]func(Expression) (Expression, error)),
		dependencies:   make([]string, 0),
	}

	p.registerPrefix(tokenizer.TokIdent, p.parseIdentifier)
	p.registerPrefix(tokenizer.TokNil, p.parseNil)
	p.registerPrefix(tokenizer.TokTrue, p.parseBoolean)
	p.registerPrefix(tokenizer.TokFalse, p.parseBoolean)
	p.registerPrefix(tokenizer.TokNumeric, p.parseNumeric)

	p.registerPrefix(tokenizer.TokString, p.parseStringLiteral)

	p.registerPrefix(tokenizer.TokNot, p.parsePrefixExpression)
	p.registerPrefix(tokenizer.TokBang, p.parsePrefixExpression)
	p.registerPrefix(tokenizer.TokLparen, p.parseGroupExpression)
	p.registerPrefix(tokenizer.TokLbrack, p.parseArrayExpression)
	p.registerPrefix(tokenizer.TokMinus, p.parsePrefixExpression)

	p.registerStdInfix(
		tokenizer.TokAnd,
		tokenizer.TokEq,
		tokenizer.TokNeq,
		tokenizer.TokLt,
		tokenizer.TokLte,
		tokenizer.TokGt,
		tokenizer.TokGte,
		tokenizer.TokPlus,
		tokenizer.TokMinus,
		tokenizer.TokAsterisk,
		tokenizer.TokSlash,
		tokenizer.TokIn,
		tokenizer.TokPower,
		tokenizer.TokModulo,
		tokenizer.TokNot,
		tokenizer.TokOr,
		tokenizer.TokElvis,
	)

	p.registerInfix(tokenizer.TokDot, p.parseChain)
	p.registerInfix(tokenizer.TokOptionalChain, p.parseChain)
	p.registerInfix(tokenizer.TokLparen, p.parseFunctionCall)
	p.registerInfix(tokenizer.TokLbrack, p.parsePropertyAccess)
	p.registerInfix(tokenizer.TokQuestion, p.parseTernary)

	return p
}

func (p *parser) parse() (*WrappedExpression, error) {
	p.advanceToken()

	expr, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}

	if p.nextToken != nil {
		return nil, p.error(
			fmt.Sprintf("unexpected token `%s`", p.nextToken.Literal),
			p.nextToken.Location,
		)
	}

	return &WrappedExpression{
		Expr:         expr,
		Dependencies: p.dependencies,
	}, nil
}

func (p *parser) parseExpression(precedence Precedence) (Expression, error) {
	if p.currentToken == nil {
		return nil, p.error("unexpected end of statement", p.previousToken.Location.PointAfter())
	}

	prefix := p.prefixParseFns[p.currentToken.Kind]
	if prefix == nil {
		return nil, p.error("unexpected token "+p.currentToken.Literal, p.currentToken.Location)
	}

	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}
	for p.currentToken != nil && precedence < p.nextPrecedence() {
		infix := p.infixParseFns[p.nextToken.Kind]
		if infix == nil {
			return leftExp, nil
		}

		p.advanceToken()

		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}

	return leftExp, nil
}

func (p *parser) parseGroupExpression() (Expression, error) {
	p.advanceToken()
	exp, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	if !p.expectNext(tokenizer.TokRparen) {
		return nil, p.error("unclosed parenthesis", p.nextToken.Location)
	}

	return exp, nil
}

func (p *parser) parsePrefixExpression() (Expression, error) {
	expr := &PrefixExpression{
		Token: p.currentToken,
		Op:    p.currentToken.Kind,
	}
	p.advanceToken()
	var err error
	expr.Right, err = p.parseExpression(PrecPrefix)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parseInfixExpression(left Expression) (Expression, error) {
	var err error

	currentOperand := p.currentToken.Literal
	if currentOperand == "not" {
		nextKind := p.nextToken.Kind
		if nextKind != tokenizer.TokIn {
			return nil, p.error(fmt.Sprintf("unexpected infix negation `not %s`, expected `not in`", p.nextToken.Literal), p.nextToken.Location)
		}
	}

	if p.currentToken.Literal == "not" {
		expr := &PrefixExpression{
			Token: p.currentToken,
			Op:    "not",
		}
		p.advanceToken()
		expr.Right = &InfixExpression{
			Token: p.currentToken,
			Op:    p.currentToken.Kind,
			Left:  left,
		}
		precedence := p.currentPrecedence()
		p.advanceToken()
		expr.Right.(*InfixExpression).Right, err = p.parseExpression(precedence)
		return expr, err
	} else {
		expr := &InfixExpression{
			Token: p.currentToken,
			Op:    p.currentToken.Kind,
			Left:  left,
		}
		precedence := p.currentPrecedence()
		p.advanceToken()
		expr.Right, err = p.parseExpression(precedence)
		return expr, err
	}
}

func (p *parser) parseTernary(left Expression) (Expression, error) {
	var err error
	expr := &Ternary{
		Token:     p.currentToken,
		Condition: left,
	}
	p.advanceToken()
	expr.Consequence, err = p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	if !p.expectNext(tokenizer.TokColon) {
		return nil, p.error("expected `:`", p.nextToken.Location)
	}
	p.advanceToken()
	expr.Alternative, err = p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parseChain(left Expression) (Expression, error) {
	expr := &Chain{
		Token:      p.currentToken,
		Left:       left,
		IsOptional: p.currentToken.Kind == tokenizer.TokOptionalChain,
	}

	if !p.expectNext(tokenizer.TokIdent) {
		return nil, p.error(fmt.Sprintf("unexpected \"%s\", expected \"identifier\"", p.nextToken.Kind), p.nextToken.Location)
	}

	expr.Right = &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	return expr, nil
}

func (p *parser) parseArrayExpression() (Expression, error) {
	var err error
	array := &Array{Token: p.currentToken}
	array.Items, err = p.parseExpressionList(tokenizer.TokRbrack)
	return array, err
}

func (p *parser) parsePropertyAccess(left Expression) (Expression, error) {
	var err error
	arr := &FieldAccess{
		Token:    p.currentToken,
		Accessed: left,
	}
	p.advanceToken()
	arr.Index, err = p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	if !p.nextIs(tokenizer.TokRbrack) {
		return nil, p.error("unclosed property access", p.nextToken.Location)
	}
	p.advanceToken()
	return arr, nil
}

func (p *parser) parseExpressionList(end string) ([]Expression, error) {
	list := make([]Expression, 0)
	if p.nextIs(end) {
		p.advanceToken()
		return list, nil
	}

	p.advanceToken()
	listElement, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	list = append(list, listElement)

	for p.nextIs("comma") {
		p.advanceToken()
		p.advanceToken()
		listElement, err = p.parseExpression(PrecLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, listElement)
	}

	if !p.expectNext(end) {
		location := p.currentToken.Location
		if p.nextToken != nil {
			location = p.nextToken.Location
		}
		return nil, p.error("unclosed list", location)
	}

	return list, nil
}

func (p *parser) parseIdentifier() (Expression, error) {
	if !slices.Contains(builtinNames, p.currentToken.Literal) {
		p.dependencies = append(p.dependencies, p.currentToken.Literal)
	}
	return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *parser) parseFunctionCall(left Expression) (Expression, error) {
	if ident, ok := left.(*Identifier); ok {
		functionName := ident.Value
		if slices.Index(builtinNames, functionName) != -1 {
			callToken := p.currentToken
			arguments, err := p.parseExpressionList(tokenizer.TokRparen)
			if err != nil {
				return nil, err
			}
			return &Builtin{
				location: callToken.Location,
				Token:    callToken,
				Name:     functionName,
				Args:     arguments,
			}, nil
		}
	}

	var callToken = p.currentToken

	arguments, err := p.parseExpressionList(tokenizer.TokRparen)
	if err != nil {
		return nil, err
	}

	return &FunctionCall{
		Token:  callToken,
		Called: left,
		Args:   arguments,
	}, nil
}

func (p *parser) parseStringLiteral() (Expression, error) {
	return &StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *parser) parseBoolean() (Expression, error) {
	return &Boolean{Token: p.currentToken, Value: p.currentToken.Kind == tokenizer.TokTrue}, nil
}

func (p *parser) parseNil() (Expression, error) {
	return &Nil{Token: p.currentToken}, nil
}

func (p *parser) registerPrefix(forKind string, fn func() (Expression, error)) {
	p.prefixParseFns[forKind] = fn
}

func (p *parser) registerInfix(forKind string, fn func(left Expression) (Expression, error)) {
	p.infixParseFns[forKind] = fn
}

func (p *parser) registerStdInfix(forKinds ...string) {
	for _, kind := range forKinds {
		p.infixParseFns[kind] = p.parseInfixExpression
	}
}

func (p *parser) currentPrecedence() Precedence {
	if precedence, ok := precedences[p.currentToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *parser) nextPrecedence() Precedence {
	if p.nextToken == nil {
		return PrecLowest
	}

	if precedence, ok := precedences[p.nextToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *parser) currentIs(t string) bool {
	return p.currentToken.Kind == t
}

func (p *parser) nextIs(kinds ...string) bool {
	if p.nextToken == nil {
		return false
	}
	found := false
	for _, kind := range kinds {
		if kind == p.nextToken.Kind {
			found = true
			break
		}
	}
	return found
}

func (p *parser) nextIsEnd() bool {
	return p.nextToken == nil
}

func (p *parser) expectNext(kind string) bool {
	if p.nextToken == nil {
		return false
	}
	if p.nextToken.Kind == kind {
		p.advanceToken()
		return true
	}
	return false
}

func (p *parser) advanceToken() {
	p.previousToken = p.currentToken

	p.cursor++
	if p.cursor >= len(p.tokens) {
		p.currentToken = nil
		p.nextToken = nil
		return
	}

	p.currentToken = &p.tokens[p.cursor]

	if p.cursor+1 >= len(p.tokens) {
		p.nextToken = nil
	} else {
		p.nextToken = &p.tokens[p.cursor+1]
	}
}

func (p *parser) error(message string, location helpers.Location) error {
	return errors2.New(message, p.file.Name, p.file.Content, location, location.FromOther())
}

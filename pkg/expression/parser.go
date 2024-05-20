package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"slices"
	"strconv"
	"strings"
)

type _parser struct {
	previousToken  *tokenizer.Token
	currentToken   *tokenizer.Token
	nextToken      *tokenizer.Token
	tokens         []tokenizer.Token
	prefixParseFns map[string]func() (Expression, error)
	infixParseFns  map[string]func(Expression) (Expression, error)
	cursor         int
	dependencies   []string
	chain          bool
}

type Precedence int

const (
	_ Precedence = iota
	PrecLowest
	PrecOr
	PrecAnd
	PrecElvis
	PrecEqual
	PrecLessGreater
	PrecInclusion
	PrecInfix
	PrecMultiply
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

	tokenizer.TokLt:  PrecLessGreater,
	tokenizer.TokLte: PrecLessGreater,
	tokenizer.TokGt:  PrecLessGreater,
	tokenizer.TokGte: PrecLessGreater,

	tokenizer.TokIn: PrecInclusion,

	tokenizer.TokNot:   PrecInfix,
	tokenizer.TokPlus:  PrecInfix,
	tokenizer.TokMinus: PrecInfix,

	tokenizer.TokDot:           PrecChain,
	tokenizer.TokOptionalChain: PrecChain,

	tokenizer.TokLparen: PrecCall,
	tokenizer.TokLbrack: PrecCall,

	tokenizer.TokAsterisk: PrecMultiply,
	tokenizer.TokSlash:    PrecMultiply,
	tokenizer.TokModulo:   PrecMultiply,
	tokenizer.TokFloorDiv: PrecMultiply,

	tokenizer.TokPower: PrecPower,

	tokenizer.TokBang: PrecPrefix,
}

func Parse(tokens []tokenizer.Token) (*WrappedExpression, error) {
	p := newParser(tokens)
	return p.parser()
}

func newParser(tokens []tokenizer.Token) *_parser {
	p := &_parser{
		cursor:         -1,
		tokens:         tokens,
		prefixParseFns: make(map[string]func() (Expression, error)),
		infixParseFns:  make(map[string]func(Expression) (Expression, error)),
		dependencies:   make([]string, 0),
		chain:          false,
	}

	p.registerPrefix(tokenizer.TokIdent, p.parseIdentifier)
	p.registerPrefix(tokenizer.TokNil, p.parseNil)
	p.registerPrefix(tokenizer.TokTrue, p.parseBoolean)
	p.registerPrefix(tokenizer.TokFalse, p.parseBoolean)
	p.registerPrefix(tokenizer.TokNumber, p.parseNumeric)
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
		tokenizer.TokFloorDiv,
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

func (p *_parser) parser() (*WrappedExpression, error) {
	p.advanceToken()

	expr, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}

	if p.nextToken != nil {
		return nil, errors2.New(
			fmt.Sprintf("unexpected token %s", p.nextToken.Literal),
			p.nextToken.Location,
		)
	}

	return &WrappedExpression{
		Expr:         expr,
		Dependencies: p.dependencies,
	}, nil
}

func (p *_parser) parseExpression(precedence Precedence) (Expression, error) {
	if p.currentToken == nil {
		return nil, errors2.New("unexpected end of expression", p.previousToken.Location)
	}

	prefix := p.prefixParseFns[p.currentToken.Kind]
	if prefix == nil {
		return nil, errors2.New("unexpected token "+p.currentToken.Literal, p.currentToken.Location)
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

func (p *_parser) parseGroupExpression() (Expression, error) {
	p.advanceToken()
	exp, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	if !p.expectNext(tokenizer.TokRparen) {
		return nil, errors2.New("unclosed parenthesis", p.nextToken.Location)
	}

	return exp, nil
}

func (p *_parser) parsePrefixExpression() (Expression, error) {
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

func (p *_parser) parseInfixExpression(left Expression) (Expression, error) {
	var err error

	currentOperand := p.currentToken.Literal
	if currentOperand == "not" {
		nextKind := p.nextToken.Kind
		if nextKind != tokenizer.TokIn {
			return nil, errors2.New(fmt.Sprintf("unexpected infix negation `not %s`, expected `not in`", p.nextToken.Literal), p.nextToken.Location)
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

func (p *_parser) parseTernary(left Expression) (Expression, error) {
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
		return nil, errors2.New("expected `:`", p.nextToken.Location)
	}
	p.advanceToken()
	expr.Alternative, err = p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *_parser) parseChain(left Expression) (Expression, error) {
	p.chain = true
	var err error
	expr := &Chain{
		Token:      p.currentToken,
		Left:       left,
		IsOptional: p.currentToken.Kind == tokenizer.TokOptionalChain,
	}

	if !p.expectNext(tokenizer.TokIdent) {
		return nil, errors2.New(fmt.Sprintf("unexpected `%s`, expected `identifier`", p.nextToken.Kind), p.nextToken.Location)
	}

	rightIdent, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	expr.Right = rightIdent.(*Identifier)

	p.chain = false

	return expr, nil
}

func (p *_parser) parseArrayExpression() (Expression, error) {
	var err error
	array := &Array{Token: p.currentToken}
	array.Items, err = p.parseExpressionList(tokenizer.TokRbrack)
	return array, err
}

func (p *_parser) parsePropertyAccess(left Expression) (Expression, error) {
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
		return nil, errors2.New("unclosed property access", p.nextToken.Location)
	}
	p.advanceToken()
	return arr, nil
}

func (p *_parser) parseExpressionList(end string) ([]Expression, error) {
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
		return nil, errors2.New("unclosed list", location)
	}

	return list, nil
}

func (p *_parser) parseIdentifier() (Expression, error) {
	if !p.chain && !slices.Contains(builtinNames, p.currentToken.Literal) {
		p.dependencies = append(p.dependencies, p.currentToken.Literal)
	}
	return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *_parser) parseFunctionCall(left Expression) (Expression, error) {
	if left.Type() == "identifier" {
		functionName := left.(*Identifier).Value
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
	arguments, err := p.parseExpressionList(tokenizer.TokRparen)
	if err != nil {
		return nil, err
	}

	return &FunctionCall{
		Token:  p.currentToken,
		Called: left,
		Args:   arguments,
	}, nil
}

func (p *_parser) parseStringLiteral() (Expression, error) {
	return &StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *_parser) parseNumeric() (Expression, error) {
	if strings.Contains(p.currentToken.Literal, ".") {
		res, err := strconv.ParseFloat(p.currentToken.Literal, 64)
		if err != nil {
			return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as floating point", p.nextToken.Location)
		}
		return &Float{Token: p.currentToken, Value: res}, nil
	}

	res, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as integer", p.nextToken.Location)
	}

	return &Integer{Token: p.currentToken, Value: int(res)}, nil
}

func (p *_parser) parseBoolean() (Expression, error) {
	return &Boolean{Token: p.currentToken, Value: p.currentToken.Kind == tokenizer.TokTrue}, nil
}

func (p *_parser) parseNil() (Expression, error) {
	return &Nil{Token: p.currentToken}, nil
}

func (p *_parser) registerPrefix(forKind string, fn func() (Expression, error)) {
	p.prefixParseFns[forKind] = fn
}

func (p *_parser) registerInfix(forKind string, fn func(left Expression) (Expression, error)) {
	p.infixParseFns[forKind] = fn
}

func (p *_parser) registerStdInfix(forKinds ...string) {
	for _, kind := range forKinds {
		p.infixParseFns[kind] = p.parseInfixExpression
	}
}

func (p *_parser) currentPrecedence() Precedence {
	if precedence, ok := precedences[p.currentToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *_parser) nextPrecedence() Precedence {
	if p.nextToken == nil {
		return PrecLowest
	}

	if precedence, ok := precedences[p.nextToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *_parser) currentIs(t string) bool {
	return p.currentToken.Kind == t
}

func (p *_parser) nextIs(kinds ...string) bool {
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

func (p *_parser) nextIsEnd() bool {
	return p.nextToken == nil
}

func (p *_parser) expectNext(kind string) bool {
	if p.nextToken == nil {
		return false
	}
	if p.nextToken.Kind == kind {
		p.advanceToken()
		return true
	}
	return false
}

func (p *_parser) advanceToken() {
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

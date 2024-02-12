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
	requiredIdents []string
	chain          bool
}

type Precedence int

const (
	_ Precedence = iota
	PrecLowest
	PrecOr
	PrecAnd
	PrecEqual
	PrecLessGreater
	PrecInclusion
	PrecSets
	PrecInfix
	PrecMultiply
	PrecPower
	PrecPrefix
	PrecCall
	PrecChain
)

var precedences = map[string]Precedence{
	"ident": PrecLowest,

	"or":  PrecOr,
	"and": PrecAnd,

	"eq":  PrecEqual,
	"neq": PrecEqual,

	"lt":  PrecLessGreater,
	"lte": PrecLessGreater,
	"gt":  PrecLessGreater,
	"gte": PrecLessGreater,

	"in": PrecInclusion,

	"subset":       PrecSets,
	"propersubset": PrecSets,

	"not":        PrecInfix,
	"plus":       PrecInfix,
	"minus":      PrecInfix,
	"ampersand":  PrecInfix,
	"circumflex": PrecInfix,

	"dot":                      PrecChain,
	tokenizer.TokOptionalChain: PrecChain,

	tokenizer.TokLparen: PrecCall,
	tokenizer.TokLbrack: PrecCall,

	"asterisk":  PrecMultiply,
	"slash":     PrecMultiply,
	"mod":       PrecMultiply,
	"floor_div": PrecMultiply,

	"power": PrecPower,

	"bang": PrecPrefix,
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
		requiredIdents: make([]string, 0),
		chain:          false,
	}

	p.registerPrefix(tokenizer.TokIdent, p.parseIdentifier)
	p.registerPrefix(tokenizer.TokTrue, p.parseBoolean)
	p.registerPrefix(tokenizer.TokFalse, p.parseBoolean)
	p.registerPrefix(tokenizer.TokNumber, p.parseNumeric)
	p.registerPrefix(tokenizer.TokString, p.parseStringLiteral)

	p.registerPrefix(tokenizer.TokNot, p.parsePrefixExpression)
	p.registerPrefix(tokenizer.TokBang, p.parsePrefixExpression)
	p.registerPrefix(tokenizer.TokLparen, p.parseGroupExpression)
	p.registerPrefix(tokenizer.TokLbrack, p.parseArrayExpression)

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
		tokenizer.TokAmpersand,
		tokenizer.TokOr,
	)

	p.registerInfix(tokenizer.TokDot, p.parseVariableAccessExpression)
	p.registerInfix(tokenizer.TokOptionalChain, p.parseVariableAccessExpression)
	p.registerInfix(tokenizer.TokLparen, p.parseFunctionCall)
	p.registerInfix(tokenizer.TokLbrack, p.parseArrayAccessExpression)

	return p
}

func (p *_parser) parser() (*WrappedExpression, error) {
	p.advanceToken()

	expr, err := p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}

	if p.nextToken != nil {
		return nil, errors2.NewErrorWithLocation(
			fmt.Sprintf("unexpected token %s", p.nextToken.Literal),
			p.nextToken.LocationStart,
		)
	}

	return &WrappedExpression{
		Expr:           expr,
		RequiredIdents: p.requiredIdents,
	}, nil
}

func (p *_parser) parseExpression(precedence Precedence) (Expression, error) {
	if p.currentToken == nil {
		return nil, errors2.NewErrorWithLocation("unexpected end of expression", p.previousToken.LocationEnd)
	}

	prefix := p.prefixParseFns[p.currentToken.Kind]
	if prefix == nil {
		return nil, errors2.NewErrorWithLocation("unexpected token "+p.currentToken.Literal, p.currentToken.LocationStart)
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
		return nil, errors2.NewErrorWithLocation("unclosed parenthesis", p.nextToken.LocationStart)
	}

	return exp, nil
}

func (p *_parser) parsePrefixExpression() (Expression, error) {
	expr := &PrefixExpression{
		Token: p.currentToken,
		Op:    p.currentToken.Literal,
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
			return nil, errors2.NewErrorWithLocation("unexpected negation `not "+p.nextToken.Literal+"`", p.nextToken.LocationStart)
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

func (p *_parser) parseRangeExpression(left Expression) (Expression, error) {
	var err error

	expr := &Range{
		Token: p.currentToken,
		Start: left,
	}
	p.advanceToken()
	expr.End, err = p.parseExpression(PrecLowest)
	return expr, err
}

func (p *_parser) parseVariableAccessExpression(left Expression) (Expression, error) {
	p.chain = true
	var err error
	expr := &VariableAccess{
		Token:      p.currentToken,
		Left:       left,
		IsOptional: p.currentToken.Kind == tokenizer.TokOptionalChain,
	}

	p.advanceToken()

	expr.Right, err = p.parseExpression(PrecChain)
	if err != nil {
		return nil, err
	}
	for p.nextIs("dot") || p.nextIs(tokenizer.TokOptionalChain) || p.nextIs(tokenizer.TokLbrack) {
		p.advanceToken()

		if p.currentIs(tokenizer.TokLbrack) {
			return p.parseArrayAccessExpression(expr)
		}

		expr.Right, err = p.parseVariableAccessExpression(expr)
		if err != nil {
			return nil, err
		}
	}

	p.chain = false

	return expr, nil
}

func (p *_parser) parseArrayExpression() (Expression, error) {
	var err error
	array := &Array{Token: p.currentToken}
	array.Items, err = p.parseExpressionList(tokenizer.TokRbrack)
	return array, err
}

func (p *_parser) parseArrayAccessExpression(left Expression) (Expression, error) {
	var err error
	arr := &ArrayAccess{
		Token:    p.currentToken,
		Accessed: left,
	}
	p.advanceToken()
	arr.Index, err = p.parseExpression(PrecLowest)
	if err != nil {
		return nil, err
	}
	if !p.nextIs(tokenizer.TokRbrack) {
		return nil, errors2.NewErrorWithLocation("unclosed array access", p.nextToken.LocationStart)
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
		return nil, errors2.NewErrorWithLocation("unclosed list", p.nextToken.LocationStart)
	}

	return list, nil
}

func (p *_parser) parseIdentifier() (Expression, error) {
	if !p.chain && !slices.Contains(builtinNames, p.currentToken.Literal) {
		p.requiredIdents = append(p.requiredIdents, p.currentToken.Literal)
	}
	return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *_parser) parseFunctionCall(left Expression) (Expression, error) {
	if left.Type() == "identifier" {
		functionName := left.(*Identifier).Value
		if slices.Index(builtinNames, functionName) != -1 {
			arguments, err := p.parseExpressionList(tokenizer.TokRparen)
			if err != nil {
				return nil, err
			}

			return &Builtin{
				Token: p.currentToken,
				Name:  functionName,
				Args:  arguments,
			}, nil
		}
	}
	arguments, err := p.parseExpressionList(tokenizer.TokRparen)
	if err != nil {
		return nil, err
	}

	call := &FunctionCall{
		Token:  p.currentToken,
		Called: left,
		Args:   arguments,
	}

	return call, nil
}

func (p *_parser) parseStringLiteral() (Expression, error) {
	return &StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}, nil
}

func (p *_parser) parseNumeric() (Expression, error) {
	if strings.Contains(p.currentToken.Literal, ".") {
		res, err := strconv.ParseFloat(p.currentToken.Literal, 64)
		if err != nil {
			return nil, errors2.NewErrorWithLocation("could not parse "+p.currentToken.Literal+" as numeric", p.nextToken.LocationStart)
		}
		return &Numeric{Token: p.currentToken, Value: res}, nil
	}

	res, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
	if err != nil {
		return nil, errors2.NewErrorWithLocation("could not parse "+p.currentToken.Literal+" as numeric", p.nextToken.LocationStart)
	}

	return &Integer{Token: p.currentToken, Value: int(res)}, nil
}

func (p *_parser) parseBoolean() (Expression, error) {
	return &Boolean{Token: p.currentToken, Value: p.currentToken.Kind == tokenizer.TokTrue}, nil
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

func (p *_parser) nextIs(t string) bool {
	if p.nextToken == nil {
		return false
	}
	return p.nextToken.Kind == t
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

package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"slices"
	"strconv"
	"strings"
)

type Parser struct {
	currentToken   *tokenizer.Token
	nextToken      *tokenizer.Token
	tokens         []tokenizer.Token
	prefixParseFns map[string]func() Expression
	infixParseFns  map[string]func(Expression) Expression
	cursor         int
	errors         []Error
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

func NewParser(tokens []tokenizer.Token) *Parser {
	p := &Parser{
		cursor:         -1,
		tokens:         tokens,
		prefixParseFns: make(map[string]func() Expression),
		infixParseFns:  make(map[string]func(Expression) Expression),
		errors:         make([]Error, 0),
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

type Error struct {
	Message  string
	Location tokenizer.Location
}

type ParserErrors struct {
	Errors []Error
}

func (e *ParserErrors) Error() string {
	return fmt.Sprintf("parser errors: %v", e.Errors)
}

func (p *Parser) Parse() (expr Expression, err error) {
	p.advanceToken()

	expr = p.parseExpression(PrecLowest)

	if p.nextToken != nil {
		p.errors = append(p.errors, Error{"unexpected token " + p.nextToken.Literal, p.nextToken.Location})
	}

	if len(p.errors) > 0 {
		return nil, &ParserErrors{p.errors}
	}

	return expr, nil
}

func (p *Parser) parseExpression(precedence Precedence) Expression {
	if p.currentToken == nil {
		p.errors = append(p.errors, Error{"unexpected EOF", tokenizer.Location{Line: -1, Column: -1}})
		return nil
	}

	prefix := p.prefixParseFns[p.currentToken.Kind]
	if prefix == nil {
		p.errors = append(p.errors, Error{"no prefix parse function for " + p.currentToken.String(), p.nextToken.Location})
		return nil
	}

	leftExp := prefix()
	for p.currentToken != nil && precedence < p.nextPrecedence() {
		infix := p.infixParseFns[p.nextToken.Kind]
		if infix == nil {
			return leftExp
		}

		p.advanceToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseGroupExpression() Expression {
	p.advanceToken()
	exp := p.parseExpression(PrecLowest)
	if !p.expectNext(tokenizer.TokRparen) {
		return nil
	}
	return exp
}

func (p *Parser) parsePrefixExpression() Expression {
	expr := &PrefixExpression{
		Token: p.currentToken,
		Op:    p.currentToken.Literal,
	}
	p.advanceToken()
	expr.Right = p.parseExpression(PrecPrefix)
	return expr
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	currentOperand := p.currentToken.Literal
	if currentOperand == "not" {
		nextKind := p.nextToken.Kind
		if nextKind != tokenizer.TokIn {
			p.errors = append(p.errors, Error{"unexpected negation `not " + p.nextToken.Literal + "`", p.nextToken.Location})
			return nil
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
		expr.Right.(*InfixExpression).Right = p.parseExpression(precedence)
		return expr
	} else {
		expr := &InfixExpression{
			Token: p.currentToken,
			Op:    p.currentToken.Kind,
			Left:  left,
		}
		precedence := p.currentPrecedence()
		p.advanceToken()
		expr.Right = p.parseExpression(precedence)
		return expr
	}
}

func (p *Parser) parseRangeExpression(left Expression) Expression {
	expr := &Range{
		Token: p.currentToken,
		Start: left,
	}
	p.advanceToken()
	expr.End = p.parseExpression(PrecLowest)
	return expr
}

func (p *Parser) parseVariableAccessExpression(left Expression) Expression {
	expr := &VariableAccess{
		Token:      p.currentToken,
		Left:       left,
		IsOptional: p.currentToken.Kind == tokenizer.TokOptionalChain,
	}

	p.advanceToken()

	expr.Right = p.parseExpression(PrecChain)
	for p.nextIs("dot") || p.nextIs(tokenizer.TokOptionalChain) || p.nextIs(tokenizer.TokLbrack) {
		p.advanceToken()

		if p.currentIs(tokenizer.TokLbrack) {
			return p.parseArrayAccessExpression(expr)
		}

		expr.Right = p.parseVariableAccessExpression(expr)
	}

	return expr
}

func (p *Parser) parseArrayExpression() Expression {
	array := &Array{Token: p.currentToken}
	array.Items = p.parseExpressionList(tokenizer.TokRbrack)
	return array
}

func (p *Parser) parseArrayAccessExpression(left Expression) Expression {
	arr := &ArrayAccess{
		Token:    p.currentToken,
		Accessed: left,
	}
	p.advanceToken()
	arr.Index = p.parseExpression(PrecLowest)
	if !p.nextIs(tokenizer.TokRbrack) {
		return nil
	}
	p.advanceToken()
	return arr
}

func (p *Parser) parseExpressionList(end string) []Expression {
	list := make([]Expression, 0)
	if p.nextIs(end) {
		p.advanceToken()
		return list
	}

	p.advanceToken()
	list = append(list, p.parseExpression(PrecLowest))

	for p.nextIs("comma") {
		p.advanceToken()
		p.advanceToken()
		list = append(list, p.parseExpression(PrecLowest))
	}

	if !p.expectNext(end) {
		panic("unclosed array")
		return nil
	}

	return list
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseFunctionCall(left Expression) Expression {
	if left.Type() == "identifier" {
		functionName := left.(*Identifier).Value
		if slices.Index(builtinNames, functionName) != -1 {
			return &Builtin{
				Token: p.currentToken,
				Name:  functionName,
				Args:  p.parseExpressionList(tokenizer.TokRparen),
			}
		}
	}

	call := &FunctionCall{
		Token:  p.currentToken,
		Called: left,
		Args:   p.parseExpressionList(tokenizer.TokRparen),
	}

	return call
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseNumeric() Expression {
	if strings.Contains(p.currentToken.Literal, ".") {
		res, err := strconv.ParseFloat(p.currentToken.Literal, 64)
		if err != nil {
			p.errors = append(p.errors, Error{"could not parse " + p.currentToken.Literal + " as numeric", p.nextToken.Location})
			return nil
		}

		return &Numeric{Token: p.currentToken, Value: res}
	}

	res, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, Error{"could not parse " + p.currentToken.Literal + " as numeric", p.nextToken.Location})
		return nil
	}

	return &Integer{Token: p.currentToken, Value: int(res)}
}

func (p *Parser) parseBoolean() Expression {
	return &Boolean{Token: p.currentToken, Value: p.currentToken.Kind == tokenizer.TokTrue}
}

func (p *Parser) registerPrefix(forKind string, fn func() Expression) {
	p.prefixParseFns[forKind] = fn
}

func (p *Parser) registerInfix(forKind string, fn func(left Expression) Expression) {
	p.infixParseFns[forKind] = fn
}

func (p *Parser) registerStdInfix(forKinds ...string) {
	for _, kind := range forKinds {
		p.infixParseFns[kind] = p.parseInfixExpression
	}
}

func (p *Parser) currentPrecedence() Precedence {
	if precedence, ok := precedences[p.currentToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *Parser) nextPrecedence() Precedence {
	if p.nextToken == nil {
		return PrecLowest
	}

	if precedence, ok := precedences[p.nextToken.Kind]; ok {
		return precedence
	}
	return PrecLowest
}

func (p *Parser) currentIs(t string) bool {
	return p.currentToken.Kind == t
}

func (p *Parser) nextIs(t string) bool {
	if p.nextToken == nil {
		return false
	}
	return p.nextToken.Kind == t
}

func (p *Parser) nextIsEnd() bool {
	return p.nextToken == nil
}

func (p *Parser) expectNext(kind string) bool {
	if p.nextToken == nil {
		p.errors = append(p.errors, Error{"unexpected EOF", tokenizer.Location{-1, -1}})
		return false
	}
	if p.nextToken.Kind == kind {
		p.advanceToken()
		return true
	}
	p.errors = append(p.errors, Error{"expected " + string(kind) + ", got " + string(p.nextToken.Kind), p.nextToken.Location})
	return false
}

func (p *Parser) advanceToken() {
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

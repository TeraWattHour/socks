package parser

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
	"strings"
)

type _parser struct {
	elements []tokenizer.Element
	programs []Program
	cursor   int
	piece    tokenizer.Element
	unclosed []Statement
}

func Parse(elements []tokenizer.Element) ([]Program, error) {
	parser := &_parser{
		elements: elements,
		cursor:   -1,
		programs: make([]Program, 0),
		unclosed: make([]Statement, 0),
	}
	return parser.Parse()
}

func (p *_parser) Parse() ([]Program, error) {
	p.next()

	for p.piece != nil {
		switch p.piece.Kind() {
		case tokenizer.TextKind:
			p.programs = append(p.programs, &Text{string(p.piece.(tokenizer.Text))})
		case tokenizer.MustacheKind:
			piece := p.piece.(*tokenizer.Mustache)

			expr, err := expression.Parse(piece.Tokens)
			if err != nil {
				return nil, err
			}
			compiled, err := expression.NewCompiler(expr.Expr).Compile()
			if err != nil {
				return nil, err
			}

			vm := expression.NewVM(compiled)

			p.addDependencies(expr.Dependencies...)
			p.programs = append(p.programs, &Expression{
				Program:      vm,
				tag:          piece,
				dependencies: expr.Dependencies,
			})
		case tokenizer.StatementKind:
			statement, err := p.parseStatement()
			if err != nil {
				return nil, err
			} else if statement != nil {
				p.programs = append(p.programs, statement)
			}
		}

		p.next()
	}

	if len(p.unclosed) > 0 {
		return nil, errors.New("unclosed tag", p.unclosed[len(p.unclosed)-1].Location())
	}

	return p.programs, nil
}

func (p *_parser) parseStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)
	switch piece.Instruction {
	case "if":
		return p.parseIfStatement()
	case "for":
		return p.parseForStatement()
	case "extend":
		return p.parseExtendStatement()
	case "define":
		return p.parseDefineStatement()
	case "slot":
		return p.parseSlotStatement()
	case "template":
		return p.parseTemplateStatement()
	default:
		if strings.HasPrefix(piece.Instruction, "end") {
			return p.parseEndStatement()
		}
	}

	return nil, errors.New("unrecognised token: '@"+piece.Instruction+"'", piece.Location)
}

func (p *_parser) parseIfStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	expr, err := expression.Parse(piece.Tokens)
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr.Expr).Compile()
	if err != nil {
		return nil, err
	}

	vm := expression.NewVM(compiled)

	statement := &IfStatement{
		Program:  vm,
		location: piece.Location,
	}

	p.unclosed = append(p.unclosed, statement)
	p.addDependencies(expr.Dependencies...)

	return statement, nil
}

// ForStatement ::= "(" Identifier "in" Expression ("with" Identifier)? ")"
func (p *_parser) parseForStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	s := &ForStatement{
		location: piece.Location,
	}

	tokens := helpers.Queue[tokenizer.Token](piece.Tokens)
	if tokens.IsEmpty() {
		return nil, errors.New("unexpected end of statement, expected identifier", piece.Location)
	}
	if tokens.Peek().Kind != tokenizer.TokIdent {
		return nil, errors.New(fmt.Sprintf("unexpected token %s, expected identifier", tokens.Peek().Kind), tokens.Peek().Location)
	}

	s.ValueName = tokens.Pop().Literal

	if tokens.Pop().Kind != tokenizer.TokIn {
		return nil, errors.New(fmt.Sprintf("unexpected token %s, expected `in`", tokens.Peek().Kind), tokens.Peek().Location)
	}
	if tokens.IsEmpty() {
		return nil, errors.New("unexpected end of statement, expected expression", piece.Location)
	}

	expressionTokens := helpers.Stack[tokenizer.Token]([]tokenizer.Token{})
	for !tokens.IsEmpty() && tokens.Peek().Kind != tokenizer.TokWith {
		expressionTokens.Push(tokens.Pop())
	}

	if !tokens.IsEmpty() {
		if tokens.Pop().Kind != tokenizer.TokWith {
			return nil, errors.New(fmt.Sprintf("unexpected token %s, expected `with`", tokens.Peek().Kind), tokens.Peek().Location)
		}
		if tokens.IsEmpty() {
			return nil, errors.New("unexpected end of statement, expected identifier", piece.Location)
		}
		if tokens.Peek().Kind != tokenizer.TokIdent {
			return nil, errors.New(fmt.Sprintf("unexpected token %s, expected identifier", tokens.Peek().Kind), tokens.Peek().Location)
		}
		s.KeyName = tokens.Pop().Literal
	}

	expr, err := expression.Parse(expressionTokens)
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr.Expr).Compile()
	if err != nil {
		return nil, err
	}

	s.Iterable = expression.NewVM(compiled)

	p.unclosed = append(p.unclosed, s)
	p.addDependencies(expr.Dependencies...)

	return s, nil
}

// @extend(templateName)
func (p *_parser) parseExtendStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 1 {
		return nil, errors.New("extend statement requires one argument", piece.Location)
	}

	return &ExtendStatement{
		Template: piece.Tokens[0].Literal,
		location: piece.Location,
	}, nil
}

// @define(name)
func (p *_parser) parseDefineStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 1 {
		return nil, errors.New("define statement requires one argument", piece.Location)
	}

	if len(p.unclosed) != 0 && p.unclosed[len(p.unclosed)-1].Kind() != "template" {
		return nil, errors.New("define statements must be placed inside a template block or at the root level", piece.Location)
	}

	statement := &DefineStatement{
		Name:     piece.Tokens[0].Literal,
		location: piece.Location,
	}

	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

// @slot(name)
func (p *_parser) parseSlotStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 1 {
		return nil, errors.New("slot statement requires one argument", piece.Location)
	}

	statement := &SlotStatement{
		Name:     piece.Tokens[0].Literal,
		location: piece.Location,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

// @template(name)
func (p *_parser) parseTemplateStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 1 {
		return nil, errors.New("template statement requires one argument", piece.Location)
	}

	statement := &TemplateStatement{
		Template: piece.Tokens[0].Literal,
		location: piece.Location,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

func (p *_parser) parseEndStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	target := piece.Instruction[3:]
	if target != "define" && target != "slot" && target != "template" && target != "if" && target != "for" {
		return nil, errors.New(fmt.Sprintf("unexpected token: @%s", piece.Instruction), piece.Location)
	}

	depth := len(p.unclosed)
	if depth == 0 {
		return nil, errors.New("unexpected end tag", piece.Location)
	}

	last := p.unclosed[depth-1]
	if last.(Statement).Kind() != target {
		return nil, errors.New(fmt.Sprintf("unexpected @end%s, expected @end%s", target, last.Kind()), piece.Location)
	}

	endStatement := &EndStatement{
		ClosedStatement: last,
	}

	switch last.Kind() {
	case "if":
		ifStatement := last.(*IfStatement)
		ifStatement.EndStatement = endStatement
	case "for":
		forStatement := last.(*ForStatement)
		forStatement.EndStatement = endStatement
		p.removeDependencies(forStatement.KeyName, forStatement.ValueName)
	case "slot":
		slotStatement := last.(*SlotStatement)
		slotStatement.EndStatement = endStatement
	case "define":
		defineStatement := last.(*DefineStatement)
		defineStatement.EndStatement = endStatement
	case "template":
		templateStatement := last.(*TemplateStatement)
		templateStatement.EndStatement = endStatement
	default:
		panic("unreachable")
	}

	p.programs = append(p.programs, endStatement)
	p.unclosed = p.unclosed[:depth-1]
	return nil, nil
}

func (p *_parser) next() {
	p.cursor += 1
	if p.cursor >= len(p.elements) {
		p.piece = nil
	} else {
		p.piece = p.elements[p.cursor]
	}
}

func (p *_parser) addDependencies(dependencies ...string) {
	for _, unclosed := range p.unclosed {
		switch unclosed.Kind() {
		case "if":
			ifStatement := unclosed.(*IfStatement)
			ifStatement.dependencies = append(ifStatement.dependencies, dependencies...)
		case "for":
			forStatement := unclosed.(*ForStatement)
			forStatement.dependencies = append(forStatement.dependencies, dependencies...)
			forStatement.dependencies = slices.DeleteFunc(forStatement.dependencies, func(s string) bool {
				return s == forStatement.KeyName || s == forStatement.ValueName
			})
		}
	}
}

func (p *_parser) removeDependencies(dependencies ...string) {
	for _, unclosed := range p.unclosed {
		switch unclosed.Kind() {
		case "if":
			ifStatement := unclosed.(*IfStatement)
			ifStatement.dependencies = slices.DeleteFunc(ifStatement.dependencies, func(s string) bool {
				return slices.Contains(dependencies, s)
			})
		case "for":
			forStatement := unclosed.(*ForStatement)
			forStatement.dependencies = append(forStatement.dependencies, dependencies...)
			forStatement.dependencies = slices.DeleteFunc(forStatement.dependencies, func(s string) bool {
				return slices.Contains(dependencies, s)
			})
		}
	}
}

func (p *_parser) parent() Statement {
	if len(p.unclosed) == 0 {
		return nil
	}
	return p.unclosed[len(p.unclosed)-1]
}

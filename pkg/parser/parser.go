package parser

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/expression"
	"github.com/terawatthour/socks/pkg/tokenizer"
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
				Dependencies: expr.Dependencies,
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

func (p *_parser) parseForStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	tokens := piece.Tokens

	if len(tokens) < 3 {
		return nil, errors.New("unexpected end of statement", tokens[len(tokens)-1].Location)
	}

	var keyName string
	if tokens[0].Kind != tokenizer.TokIdent {
		return nil, errors.New(fmt.Sprintf("unexpected token %s, expected identifier", tokens[0].Kind), tokens[0].Location)
	}
	valueName := tokens[0].Literal
	locals := []string{valueName}

	if tokens[1].Kind == tokenizer.TokComma {
		if len(tokens) < 4 {
			return nil, errors.New("unexpected end of statement", tokens[len(tokens)-1].Location)
		}
		if tokens[2].Kind != tokenizer.TokIdent {
			return nil, errors.New(fmt.Sprintf("unexpected token %s, expected identifier", tokens[2].Kind), tokens[2].Location)
		}
		keyName = tokens[2].Literal
		locals = append(locals, keyName)
		if tokens[3].Kind != tokenizer.TokIn {
			return nil, errors.New(fmt.Sprintf("unexpected token %s, expected `in`", tokens[3].Kind), tokens[3].Location)
		}
		tokens = tokens[4:]
	} else {
		if tokens[1].Kind != tokenizer.TokIn {
			return nil, errors.New(fmt.Sprintf("unexpected token %s, expected `in`", tokens[1].Kind), tokens[1].Location)
		}
		tokens = tokens[2:]
	}

	expr, err := expression.Parse(tokens)
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr.Expr).Compile()
	if err != nil {
		return nil, err
	}

	vm := expression.NewVM(compiled)

	statement := &ForStatement{
		KeyName:   keyName,
		ValueName: valueName,
		Iterable:  vm,
		location:  piece.Location,
	}

	p.unclosed = append(p.unclosed, statement)
	p.addDependencies(expr.Dependencies...)

	return statement, nil
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
			ifStatement.Dependencies = append(ifStatement.Dependencies, dependencies...)
		case "for":
			forStatement := unclosed.(*ForStatement)
			forStatement.Dependencies = append(forStatement.Dependencies, dependencies...)
			forStatement.Dependencies = slices.DeleteFunc(forStatement.Dependencies, func(s string) bool {
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
			ifStatement.Dependencies = slices.DeleteFunc(ifStatement.Dependencies, func(s string) bool {
				return slices.Contains(dependencies, s)
			})
		case "for":
			forStatement := unclosed.(*ForStatement)
			forStatement.Dependencies = append(forStatement.Dependencies, dependencies...)
			forStatement.Dependencies = slices.DeleteFunc(forStatement.Dependencies, func(s string) bool {
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

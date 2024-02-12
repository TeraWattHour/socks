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
	noStatic []bool
}

func Parse(elements []tokenizer.Element) ([]Program, error) {
	parser := &_parser{
		elements: elements,
		cursor:   -1,
		noStatic: []bool{},
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
			p.programs = append(p.programs, &Text{string(p.piece.(tokenizer.Text)), p.parent()})
		case tokenizer.MustacheKind:
			piece := p.piece.(*tokenizer.Mustache)
			if piece.IsComment {
				p.next()
				continue
			}

			expr, err := expression.Parse(piece.Tokens)
			if err != nil {
				return nil, err
			}
			compiled, err := expression.NewCompiler(expr.Expr).Compile()
			if err != nil {
				return nil, err
			}

			vm := expression.NewVM(compiled)

			if expr != nil {
				p.addDependencies(expr.RequiredIdents...)
				p.programs = append(p.programs, &Expression{
					Program:      vm,
					tag:          piece,
					noStatic:     p.checkNoStatic(),
					Parent:       p.parent(),
					Dependencies: expr.RequiredIdents,
				})
			}
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
		return nil, errors.NewErrorWithLocation("unclosed tag", p.unclosed[len(p.unclosed)-1].Location())
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

	return nil, errors.NewError("unrecognised token: '@" + piece.Instruction + "'")
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

	noStatic := slices.Index(piece.Flags, "nostatic") != -1

	statement := &IfStatement{
		Program:   vm,
		noStatic:  noStatic || p.checkNoStatic(),
		bodyStart: len(p.programs) + 1,
		Parent:    p.parent(),
		location:  piece.Location,
	}

	p.noStatic = append(p.noStatic, noStatic)
	p.unclosed = append(p.unclosed, statement)
	p.addDependencies(expr.RequiredIdents...)

	return statement, nil
}

func (p *_parser) parseForStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	tokens := piece.Tokens
	tokens = tokens[1 : len(tokens)-1]

	if len(tokens) < 3 {
		return nil, errors.NewError("unexpected token: '@for'")
	}

	var keyName string
	valueName := tokens[0].Literal
	locals := []string{valueName}

	if tokens[1].Kind == tokenizer.TokComma {
		if tokens[2].Kind != tokenizer.TokIdent {
			return nil, errors.NewError("unexpected token: '@for'")
		}
		keyName = tokens[2].Literal
		locals = append(locals, keyName)
		if tokens[3].Kind != tokenizer.TokIn {
			return nil, errors.NewError("unexpected token: '@for'")
		}
		tokens = tokens[4:]
	} else {
		if tokens[1].Kind != tokenizer.TokIn {
			return nil, errors.NewError("unexpected token: '@for'")
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

	noStatic := slices.Index(piece.Flags, "nostatic") != -1

	statement := &ForStatement{
		KeyName:   keyName,
		ValueName: valueName,
		Iterable:  vm,
		noStatic:  noStatic || p.checkNoStatic(),
		bodyStart: len(p.programs) + 1,
		Parent:    p.parent(),
		location:  piece.Location,
	}

	p.unclosed = append(p.unclosed, statement)
	p.noStatic = append(p.noStatic, noStatic)
	p.addDependencies(expr.RequiredIdents...)

	return statement, nil
}

// @extend(templateName)
func (p *_parser) parseExtendStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 3 {
		return nil, errors.NewError("unexpected statement: '@extend', expected template name")
	}

	return &ExtendStatement{
		Template: piece.Tokens[1].Literal,
		location: piece.Location,
	}, nil
}

// @define(name)
func (p *_parser) parseDefineStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 3 {
		return nil, errors.NewError("unexpected statement: '@define', expected name")
	}

	if len(p.unclosed) != 0 && p.unclosed[len(p.unclosed)-1].Kind() != "template" {
		return nil, errors.NewError("@define statements must be placed inside a @template block or at the root level")
	}

	statement := &DefineStatement{
		Name:      piece.Tokens[1].Literal,
		Parent:    p.parent(),
		bodyStart: len(p.programs) + 1,
		Depth:     len(p.unclosed),
		location:  piece.Location,
	}

	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

// @slot(name)
func (p *_parser) parseSlotStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 3 {
		return nil, errors.NewError("unexpected statement: '@slot', expected name")
	}

	statement := &SlotStatement{
		Name:      piece.Tokens[1].Literal,
		Parent:    p.parent(),
		bodyStart: len(p.programs) + 1,
		Depth:     len(p.unclosed),
		location:  piece.Location,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

// @template(name)
func (p *_parser) parseTemplateStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens) != 3 {
		return nil, errors.NewError("unexpected statement: '@slot', expected name")
	}

	statement := &TemplateStatement{
		Template:  piece.Tokens[1].Literal,
		BodyStart: len(p.programs) + 1,
		Parent:    p.parent(),
		Depth:     len(p.unclosed),
		location:  piece.Location,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

func (p *_parser) parseEndStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	target := piece.Instruction[3:]
	if target != "define" && target != "slot" && target != "template" && target != "if" && target != "for" {
		return nil, errors.NewError("unexpected token: '@" + piece.Instruction + "'")
	}

	depth := len(p.unclosed)
	if depth == 0 {
		return nil, errors.NewError("unexpected end tag")
	}

	last := p.unclosed[depth-1]
	if last.(Statement).Kind() != target {
		return nil, errors.NewErrorWithLocation(fmt.Sprintf("unexpected `@end%s`, expected `@end%s`", target, last.Kind()), piece.Location)
	}

	switch last.Kind() {
	case "if":
		ifStatement := last.(*IfStatement)
		ifStatement.Programs = len(p.programs) - ifStatement.bodyStart
		p.noStatic = p.noStatic[:len(p.noStatic)-1]
	case "for":
		forStatement := last.(*ForStatement)
		forStatement.Programs = len(p.programs) - forStatement.bodyStart
		p.removeDependencies(forStatement.KeyName, forStatement.ValueName)
		p.noStatic = p.noStatic[:len(p.noStatic)-1]
	case "slot":
		slotStatement := last.(*SlotStatement)
		slotStatement.Programs = len(p.programs) - slotStatement.bodyStart
	case "define":
		defineStatement := last.(*DefineStatement)
		defineStatement.Programs = len(p.programs) - defineStatement.bodyStart
	case "template":
		templateStatement := last.(*TemplateStatement)
		templateStatement.Programs = len(p.programs) - templateStatement.BodyStart
	default:
		panic("unreachable")
	}

	p.unclosed = p.unclosed[:depth-1]
	return nil, nil
}

func (p *_parser) checkNoStatic() bool {
	for _, noStatic := range p.noStatic {
		if noStatic {
			return true
		}
	}
	return false
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

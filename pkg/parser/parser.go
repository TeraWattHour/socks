package parser

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/expression"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"reflect"
	"slices"
	"strings"
)

type Parser struct {
	Tokenizer *tokenizer.Tokenizer
	Programs  []Program
	cursor    int
	piece     tokenizer.Element
	unclosed  []Statement
	noStatic  []bool
	errors    []errors.Error
}

func NewParser(tokenizer *tokenizer.Tokenizer) *Parser {
	return &Parser{
		Tokenizer: tokenizer,
		cursor:    -1,
		noStatic:  []bool{},
		Programs:  make([]Program, 0),
		unclosed:  make([]Statement, 0),
		errors:    make([]errors.Error, 0),
	}
}

func (p *Parser) Parse() ([]Program, error) {
	p.Next()

	for p.piece != nil {
		switch p.piece.Kind() {
		case "text":
			p.Programs = append(p.Programs, Text(p.piece.(tokenizer.Text)))
		case "tag":
			piece := p.piece.(*tokenizer.Tag)
			if piece.Type == tokenizer.CommentKind {
				p.Next()
				continue
			}

			expr, err := expression.NewParser(p.piece.Tokens()).Parse()
			if err != nil {
				return nil, err
			}
			compiled, err := expression.NewCompiler(expr).Compile()
			if err != nil {
				return nil, err
			}

			vm := expression.NewVM(compiled)

			if expr != nil {
				p.Programs = append(p.Programs, &PrintStatement{Program: vm, tag: piece, noStatic: p.checkNoStatic()})
			}
		case "statement":
			statement, err := p.parseStatement()
			if err != nil {
				return nil, err
			} else if statement != nil {
				p.Programs = append(p.Programs, statement)
			}
		}

		p.Next()
	}

	if len(p.unclosed) > 0 {
		return nil, errors.NewError("unclosed tag")
	}

	return p.Programs, nil
}

func (p *Parser) parseStatement() (Statement, error) {
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

func (p *Parser) parseIfStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	expr, err := expression.NewParser(piece.Tokens()).Parse()
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr).Compile()
	if err != nil {
		return nil, err
	}

	vm := expression.NewVM(compiled)

	noStatic := slices.Index(piece.Flags, "nostatic") != -1

	statement := &IfStatement{
		Program:   vm,
		noStatic:  noStatic || p.checkNoStatic(),
		bodyStart: len(p.Programs) + 1,
	}

	p.noStatic = append(p.noStatic, noStatic)
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

func (p *Parser) parseForStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	tokens := piece.Tokens()
	tokens = tokens[1 : len(tokens)-1]

	if len(tokens) < 3 {
		return nil, errors.NewError("unexpected token: '@for'")
	}

	var keyName string
	valueName := tokens[0].Literal
	if tokens[1].Kind == tokenizer.TokComma {
		if tokens[2].Kind != tokenizer.TokIdent {
			return nil, errors.NewError("unexpected token: '@for'")
		}
		keyName = tokens[2].Literal
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

	expr, err := expression.NewParser(tokens).Parse()
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr).Compile()
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
		bodyStart: len(p.Programs) + 1,
	}

	p.unclosed = append(p.unclosed, statement)
	p.noStatic = append(p.noStatic, noStatic)
	return statement, nil
}

// @extend(templateName)
func (p *Parser) parseExtendStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens()) != 3 {
		return nil, errors.NewError("unexpected statement: '@extend', expected template name")
	}

	return &ExtendStatement{
		Template: piece.Tokens()[1].Literal,
	}, nil
}

// @define(name)
func (p *Parser) parseDefineStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens()) != 3 {
		return nil, errors.NewError("unexpected statement: '@define', expected name")
	}

	if len(p.unclosed) != 0 && p.unclosed[len(p.unclosed)-1].Kind() != "template" {
		return nil, errors.NewError("@define statements must be placed inside a @template block or at the root level")
	}

	statement := &DefineStatement{
		Name:      piece.Tokens()[1].Literal,
		Parents:   p.unclosed,
		bodyStart: len(p.Programs) + 1,
	}
	p.unclosed = append(p.unclosed, statement)

	return statement, nil
}

// @slot(name)
func (p *Parser) parseSlotStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens()) != 3 {
		return nil, errors.NewError("unexpected statement: '@slot', expected name")
	}

	statement := &SlotStatement{
		Name:      piece.Tokens()[1].Literal,
		Parents:   p.unclosed,
		bodyStart: len(p.Programs) + 1,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

// @template(name)
func (p *Parser) parseTemplateStatement() (Statement, error) {
	piece := p.piece.(*tokenizer.Statement)

	if len(piece.Tokens()) != 3 {
		return nil, errors.NewError("unexpected statement: '@slot', expected name")
	}

	statement := &TemplateStatement{
		Template:  piece.Tokens()[1].Literal,
		BodyStart: len(p.Programs) + 1,
	}
	p.unclosed = append(p.unclosed, statement)
	return statement, nil
}

func (p *Parser) parseEndStatement() (Statement, error) {
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
		return nil, errors.NewError(fmt.Sprintf("unexpected end tag, no '%s' statement to close", target))
	}

	switch last.Kind() {
	case "if":
		ifStatement := last.(*IfStatement)
		ifStatement.Programs = len(p.Programs) - ifStatement.bodyStart
		p.noStatic = p.noStatic[:len(p.noStatic)-1]
	case "for":
		forStatement := last.(*ForStatement)
		forStatement.Programs = len(p.Programs) - forStatement.bodyStart
		p.noStatic = p.noStatic[:len(p.noStatic)-1]
	case "slot":
		slotStatement := last.(*SlotStatement)
		slotStatement.Programs = len(p.Programs) - slotStatement.bodyStart
	case "define":
		defineStatement := last.(*DefineStatement)
		defineStatement.Programs = len(p.Programs) - defineStatement.bodyStart
	case "template":
		templateStatement := last.(*TemplateStatement)
		templateStatement.Programs = len(p.Programs) - templateStatement.BodyStart
	}

	p.unclosed = p.unclosed[:depth-1]
	return nil, nil
}

func (p *Parser) checkNoStatic() bool {
	for _, noStatic := range p.noStatic {
		if noStatic {
			return true
		}
	}
	return false
}

func (p *Parser) Next() {
	p.cursor += 1
	if p.cursor >= len(p.Tokenizer.Elements) {
		p.piece = nil
	} else {
		p.piece = p.Tokenizer.Elements[p.cursor]
	}
}

func PrintPrograms(programs []Program) {
	fmt.Println("Programs:")
	indents := []int{}
	for _, program := range programs {
		if len(indents) > 0 {
			fmt.Print(strings.Repeat("  ", len(indents)-1), "└─")
			indents[len(indents)-1] -= 1
			if indents[len(indents)-1] == 0 {
				indents = indents[:len(indents)-1]
			}
		}
		fmt.Print(program)
		if reflect.TypeOf(program).Kind() == reflect.String {
			fmt.Println()
			continue
		}
		programsField := reflect.Indirect(reflect.ValueOf(program)).FieldByName("Programs")
		if programsField.IsValid() {
			indents = append(indents, int(programsField.Int()))
			fmt.Printf(" Programs: (%d)", int(programsField.Int()))
		}
		fmt.Println()
	}
	fmt.Println("End Programs")
}

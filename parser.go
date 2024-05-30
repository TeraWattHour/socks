package socks

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
	"strings"
)

type parser struct {
	file helpers.File

	elements helpers.Queue[tokenizer.Element]
	programs []Statement
	unclosed []Statement
}

func Parse(file helpers.File, elements []tokenizer.Element) ([]Statement, error) {
	parser := &parser{
		file:     file,
		elements: elements,
		programs: make([]Statement, 0),
		unclosed: make([]Statement, 0),
	}
	return parser.Parse()
}

func (p *parser) Parse() ([]Statement, error) {
	for !p.elements.IsEmpty() {
		switch element := p.elements.Pop().(type) {
		case tokenizer.Text:
			p.programs = append(p.programs, &Text{string(element)})
		case *tokenizer.Mustache:
			expr, err := expression.Parse(p.file, element.Tokens)
			if err != nil {
				return nil, err
			}
			compiled, err := expression.NewCompiler(expr.Expr).Compile()
			if err != nil {
				return nil, err
			}

			p.addDependencies(expr.Dependencies...)
			p.programs = append(p.programs, &Expression{
				Program:      expression.NewVM(compiled),
				tag:          element,
				dependencies: expr.Dependencies,
			})
		case *tokenizer.Statement:
			if statement, err := p.parseStatement(element); err != nil {
				return nil, err
			} else if statement != nil {
				if statement.IsClosable() {
					p.unclosed = append(p.unclosed, statement)
				}
				p.programs = append(p.programs, statement)
			}
		}
	}

	if len(p.unclosed) > 0 {
		return nil, p.error("unclosed tag", p.unclosed[len(p.unclosed)-1].Location())
	}

	return p.programs, nil
}

func (p *parser) parseStatement(statement *tokenizer.Statement) (Statement, error) {
	switch statement.Instruction {
	case "if":
		return p.parseIfStatement(statement)
	case "else":
		ifStatement, ok := p.parent().(*IfStatement)
		if !ok {
			return nil, p.error("unexpected else tag outside if statement", statement.Location)
		}
		st := &ElseStatement{location: statement.Location}
		ifStatement.ElseStatement = st
		return st, nil
	case "elif":
		ifStatement, ok := p.parent().(*IfStatement)
		if !ok {
			return nil, p.error("unexpected elif tag outside if statement", statement.Location)
		}
		if ifStatement.ElseStatement != nil {
			return nil, p.error("unexpected elif tag after else tag", statement.Location)
		}
		ast, err := expression.Parse(p.file, statement.Tokens)
		if err != nil {
			return nil, err
		}
		compiled, err := expression.NewCompiler(ast.Expr).Compile()
		if err != nil {
			return nil, err
		}

		st := &ElifStatement{
			location: statement.Location,
			Program:  expression.NewVM(compiled),
		}
		p.addDependencies(ast.Dependencies...)
		ifStatement.ElifStatements = append(ifStatement.ElifStatements, st)
		return st, nil
	case "for":
		return p.parseForStatement(statement)
	case "extend":
		return p.parseExtendStatement(statement)
	case "define":
		return p.parseDefineStatement(statement)
	case "slot":
		return p.parseSlotStatement(statement)
	case "template":
		return p.parseTemplateStatement(statement)
	default:
		if strings.HasPrefix(statement.Instruction, "end") {
			return p.parseEndStatement(statement)
		}
	}

	return nil, p.error("unrecognised token: '@"+statement.Instruction+"'", statement.Location)
}

func (p *parser) parseIfStatement(s *tokenizer.Statement) (Statement, error) {
	expr, err := expression.Parse(p.file, s.Tokens)
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr.Expr).Compile()
	if err != nil {
		return nil, err
	}

	vm := expression.NewVM(compiled)

	statement := &IfStatement{
		Program:      vm,
		location:     s.Location,
		dependencies: expr.Dependencies,
	}
	p.addDependencies(expr.Dependencies...)
	return statement, nil
}

// ForStatement ::= "(" Identifier "in" Expression ("with" Identifier)? ")"
func (p *parser) parseForStatement(s *tokenizer.Statement) (Statement, error) {
	statement := &ForStatement{
		location: s.Location,
	}

	tokens := helpers.Queue[tokenizer.Token](s.Tokens)
	if tokens.IsEmpty() {
		return nil, p.error("unexpected end of statement, expected identifier", s.Location)
	}
	if tokens.Peek().Kind != tokenizer.TokIdent {
		return nil, p.error(fmt.Sprintf("unexpected token %s, expected identifier", tokens.Peek().Kind), tokens.Peek().Location)
	}

	statement.ValueName = tokens.Pop().Literal

	if tokens.Pop().Kind != tokenizer.TokIn {
		return nil, p.error(fmt.Sprintf("unexpected token %s, expected `in`", tokens.Peek().Kind), tokens.Peek().Location)
	}
	if tokens.IsEmpty() {
		return nil, p.error("unexpected end of statement, expected expression", s.Location)
	}

	expressionTokens := helpers.Stack[tokenizer.Token]([]tokenizer.Token{})
	for !tokens.IsEmpty() && tokens.Peek().Kind != tokenizer.TokWith {
		expressionTokens.Push(tokens.Pop())
	}

	if !tokens.IsEmpty() {
		if tokens.Pop().Kind != tokenizer.TokWith {
			return nil, p.error(fmt.Sprintf("unexpected token %s, expected `with`", tokens.Peek().Kind), tokens.Peek().Location)
		}
		if tokens.IsEmpty() {
			return nil, p.error("unexpected end of statement, expected identifier", s.Location)
		}
		if tokens.Peek().Kind != tokenizer.TokIdent {
			return nil, p.error(fmt.Sprintf("unexpected token %s, expected identifier", tokens.Peek().Kind), tokens.Peek().Location)
		}
		statement.KeyName = tokens.Pop().Literal
	}

	expr, err := expression.Parse(p.file, expressionTokens)
	if err != nil {
		return nil, err
	}

	compiled, err := expression.NewCompiler(expr.Expr).Compile()
	if err != nil {
		return nil, err
	}

	statement.Iterable = expression.NewVM(compiled)
	statement.dependencies = expr.Dependencies
	p.addDependencies(expr.Dependencies...)

	return statement, nil
}

// @extend(templateName)
func (p *parser) parseExtendStatement(statement *tokenizer.Statement) (Statement, error) {
	if len(statement.Tokens) != 1 {
		return nil, p.error("extend statement requires one argument of type <string>", statement.Location)
	}

	return &ExtendStatement{
		Template: statement.Tokens[0].Literal,
		location: statement.Location,
	}, nil
}

// @define(name)
func (p *parser) parseDefineStatement(statement *tokenizer.Statement) (Statement, error) {
	if len(statement.Tokens) != 1 {
		return nil, p.error("define statement requires one argument of type <string>", statement.Location)
	}

	if len(p.unclosed) != 0 && p.unclosed[len(p.unclosed)-1].Kind() != "template" {
		return nil, p.error("define statements must be placed inside a template block or at the root level", statement.Location)
	}

	return &DefineStatement{
		Name:     statement.Tokens[0].Literal,
		location: statement.Location,
	}, nil
}

// @slot(name)
func (p *parser) parseSlotStatement(statement *tokenizer.Statement) (Statement, error) {
	if len(statement.Tokens) != 1 {
		return nil, p.error("slot statement requires one argument of type <string>", statement.Location)
	}

	return &SlotStatement{
		Name:     statement.Tokens[0].Literal,
		location: statement.Location,
	}, nil
}

// @template(name)
func (p *parser) parseTemplateStatement(statement *tokenizer.Statement) (Statement, error) {
	if len(statement.Tokens) != 1 {
		return nil, p.error("template statement requires one argument of type <string>", statement.Location)
	}

	return &TemplateStatement{
		Template: statement.Tokens[0].Literal,
		location: statement.Location,
	}, nil
}

func (p *parser) parseEndStatement(s *tokenizer.Statement) (Statement, error) {
	target := s.Instruction[3:]
	if target != "define" && target != "slot" && target != "template" && target != "if" && target != "for" {
		return nil, p.error(fmt.Sprintf("unexpected token: @%s", s.Instruction), s.Location)
	}

	depth := len(p.unclosed)
	if depth == 0 {
		return nil, p.error("unexpected end tag", s.Location)
	}

	last := p.unclosed[depth-1]
	if last.Kind() != target {
		return nil, p.error(fmt.Sprintf("unexpected `@end%s`, expected `@end%s`", target, last.Kind()), s.Location)
	}

	endStatement := &EndStatement{
		ClosedStatement: last,
	}

	switch statement := last.(type) {
	case *IfStatement:
		statement.EndStatement = endStatement
	case *ForStatement:
		statement.EndStatement = endStatement
		p.removeDependencies(statement.KeyName, statement.ValueName)
	case *SlotStatement:
		statement.EndStatement = endStatement
	case *DefineStatement:
		statement.EndStatement = endStatement
	case *TemplateStatement:
		statement.EndStatement = endStatement
	default:
		panic("unreachable, unknown closed statement type")
	}

	p.programs = append(p.programs, endStatement)
	p.unclosed = p.unclosed[:depth-1]
	return nil, nil
}

func (p *parser) addDependencies(dependencies ...string) {
	for _, unclosed := range p.unclosed {
		switch statement := unclosed.(type) {
		case *IfStatement:
			statement.dependencies = append(statement.dependencies, dependencies...)
		case *ForStatement:
			statement.dependencies = append(statement.dependencies, dependencies...)
			statement.dependencies = slices.DeleteFunc(statement.dependencies, func(s string) bool {
				return s == statement.KeyName || s == statement.ValueName
			})
		}
	}
}

func (p *parser) removeDependencies(dependencies ...string) {
	for _, unclosed := range p.unclosed {
		switch statement := unclosed.(type) {
		case *IfStatement:
			statement.dependencies = slices.DeleteFunc(statement.dependencies, func(s string) bool {
				return slices.Contains(dependencies, s)
			})
		case *ForStatement:
			statement.dependencies = append(statement.dependencies, dependencies...)
			statement.dependencies = slices.DeleteFunc(statement.dependencies, func(s string) bool {
				return slices.Contains(dependencies, s)
			})
		}
	}
}

func (p *parser) parent() Statement {
	if len(p.unclosed) == 0 {
		return nil
	}
	return p.unclosed[len(p.unclosed)-1]
}

func (p *parser) error(message string, l helpers.Location) error {
	return errors.New(message, p.file.Name, p.file.Content, l, l.FromOther())
}

package parser

import (
	"github.com/antonmedv/expr"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

type TagProgram struct {
	Tag       *tokenizer.Tag
	Statement Statement
}

type Parser struct {
	Tokenizer *tokenizer.Tokenizer
	Programs  []TagProgram
	cursor    int
	tag       *tokenizer.Tag
	unclosed  []Statement
}

type TagParser struct {
	parser    *Parser
	tag       *tokenizer.Tag
	cursor    int
	token     *tokenizer.Token
	nextToken *tokenizer.Token
}

func NewParser(tokenizer *tokenizer.Tokenizer) *Parser {
	return &Parser{
		Tokenizer: tokenizer,
		cursor:    -1,
		unclosed:  make([]Statement, 0),
	}
}

func NewTagParser(parser *Parser, tag *tokenizer.Tag) *TagParser {
	return &TagParser{
		parser: parser,
		tag:    tag,
		cursor: -1,
	}
}

func (p *Parser) Parse() error {
	p.Next()

	var programs []TagProgram

	for p.tag != nil {
		statement, err := NewTagParser(p, p.tag).Parse()
		if err != nil {
			return err
		}

		programs = append(programs, TagProgram{
			Tag:       p.tag,
			Statement: statement,
		})

		p.Next()
	}

	if len(p.unclosed) > 0 {
		tag := p.unclosed[0].Tag()
		if tag == nil {
			return errors.NewParserError("unclosed tag", -1, -1)
		}
		return errors.NewParserError("unclosed tag", tag.Start, tag.End)
	}

	p.Programs = programs
	return nil
}

// Parse parses a tag, returns a statement (Statement) that can be evaluated and an error (*Error).
// This function is called individually for each tag.
func (tp *TagParser) Parse() (st Statement, err error) {
	tp.Next()

	if tp.token == nil {
		return nil, errors.NewParserError("empty tag", tp.tag.Start, tp.tag.Start)
	}
	if tp.tag.Kind == "preprocessor" {
		switch tp.token.Kind {
		case tokenizer.TOK_EXTEND:
			return tp.parseExtendStatement()
		case tokenizer.TOK_SLOT:
			return tp.parseSlotStatement()
		case tokenizer.TOK_DEFINE:
			return tp.parseDefineStatement()
		case tokenizer.TOK_END:
			return tp.parseEndStatement()
		case tokenizer.TOK_TEMPLATE:
			return tp.parseTemplateStatement()
		default:
			return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
		}
	} else if tp.tag.Kind == "print" {
		return tp.parseVariableStatement()
	} else if tp.tag.Kind == "execute" {
		switch tp.token.Kind {
		case tokenizer.TOK_IF:
			return tp.parseIfStatement()
		case tokenizer.TOK_FOR:
			return tp.parseForStatement()
		case tokenizer.TOK_END:
			return tp.parseEndStatement()
		}
	}

	return nil, errors.NewParserError("unexpected tag type: "+string(tp.tag.Kind), tp.tag.Start, tp.tag.End)
}

func (tp *TagParser) expectNext(kind string) bool {
	if tp.nextToken == nil {
		return false
	}

	return tp.nextToken.Kind == kind
}

func (tp *TagParser) expectEnd() bool {
	return tp.nextToken == nil
}

func (tp *TagParser) parseIfStatement() (Statement, error) {
	tp.Next()
	start := tp.token.Start
	var end int

	for tp.token != nil {
		end = tp.token.Start + tp.token.Length

		tp.Next()
	}

	literal := string(tp.parser.Tokenizer.Runes[start:end])
	program, err := expr.Compile(literal)
	if err != nil {
		return nil, errors.NewParserError("unable to compile expression: "+err.Error(), tp.tag.Start, tp.tag.End)
	}

	statement := &IfStatement{
		Program:  program,
		StartTag: tp.tag,
		parents:  tp.parser.unclosed,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseEndStatement() (Statement, error) {
	depth := len(tp.parser.unclosed)
	if depth == 0 {
		return nil, errors.NewParserError("unexpected end tag", tp.tag.Start, tp.tag.End)
	}

	last := tp.parser.unclosed[depth-1]

	switch last.Kind() {
	case "define":
		last.(*DefineStatement).EndTag = tp.tag
	case "slot":
		last.(*SlotStatement).EndTag = tp.tag
	case "for":
		forStatement := last.(*ForStatement)
		forStatement.EndTag = tp.tag
		forStatement.Body = tp.parser.Tokenizer.Runes[forStatement.StartTag.End+1 : tp.tag.Start]
	case "template":
		last.(*TemplateStatement).EndTag = tp.tag
	case "if":
		ifStatement := last.(*IfStatement)
		ifStatement.EndTag = tp.tag
		ifStatement.Body = tp.parser.Tokenizer.Runes[ifStatement.StartTag.End+1 : tp.tag.Start]
	}

	tp.parser.unclosed = tp.parser.unclosed[:depth-1]

	return &EndStatement{
		closes: last,
		tag:    tp.tag,
	}, nil
}

func (tp *TagParser) parseTemplateStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}

	statement := &TemplateStatement{
		Template: tp.token.Literal,
		StartTag: tp.tag,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseExtendStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}

	return &ExtendStatement{
		Template: tp.token.Literal,
		tag:      tp.tag,
		parents:  tp.parser.unclosed,
	}, nil
}

func (tp *TagParser) parseDefineStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}

	statement := &DefineStatement{
		Name:     tp.token.Literal,
		StartTag: tp.tag,
		Parents:  tp.parser.unclosed,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseForStatement() (Statement, error) {
	var iterable Statement
	var iteratorName string
	var valueName string

	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IDENT {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}
	iteratorName = tp.token.Literal
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_COMMA {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IDENT {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}
	valueName = tp.token.Literal
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IN {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IDENT {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}
	iterable, err := tp.parseVariableStatement()
	if err != nil {
		return nil, err
	}

	statement := &ForStatement{
		IteratorName: iteratorName,
		ValueName:    valueName,
		Iterable:     iterable,
		StartTag:     tp.tag,
		EndTag:       nil,
		parents:      tp.parser.unclosed,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseSlotStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}

	statement := &SlotStatement{
		Name:     tp.token.Literal,
		StartTag: tp.tag,
		parents:  tp.parser.unclosed,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseVariableStatement() (*VariableStatement, error) {
	start := tp.token.Start
	var end int

	for tp.token != nil {
		end = tp.token.Start + tp.token.Length

		tp.Next()
	}

	literal := string(tp.parser.Tokenizer.Runes[start:end])

	program, err := expr.Compile(literal)
	if err != nil {
		return nil, errors.NewParserError("unable to compile expression: "+err.Error(), tp.tag.Start, tp.tag.End)
	}

	return &VariableStatement{
		Program: program,
		tag:     tp.tag,
		parents: tp.parser.unclosed,
	}, nil
}

func (tp *TagParser) Next() {
	tp.cursor += 1
	if tp.cursor >= len(tp.tag.Tokens) {
		tp.token = nil
	} else {
		tp.token = &tp.tag.Tokens[tp.cursor]
	}

	if tp.cursor+1 >= len(tp.tag.Tokens) {
		tp.nextToken = nil
	} else {
		tp.nextToken = &tp.tag.Tokens[tp.cursor+1]
	}
}

func (p *Parser) Next() {
	p.cursor += 1

	if p.cursor >= len(p.Tokenizer.Tags) {
		p.tag = nil
	} else {
		p.tag = &p.Tokenizer.Tags[p.cursor]
	}
}

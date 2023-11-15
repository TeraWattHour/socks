package parser

import (
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"strconv"
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
	depth     int
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
		default:
			return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
		}
	} else if tp.tag.Kind == "print" {
		switch tp.token.Kind {
		case tokenizer.TOK_DOT, tokenizer.TOK_IDENT:
			return tp.parseVariableStatement()
		case tokenizer.TOK_INTEGER:
			return tp.parseNumericStatement()
		case tokenizer.TOK_STRING:
			return tp.parseStringStatement()
		default:
			return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
		}
	} else if tp.tag.Kind == "execute" {
		switch tp.token.Kind {
		case tokenizer.TOK_FOR:
			return tp.parseForStatement()
		case tokenizer.TOK_END:
			return tp.parseEndStatement()
		}
	}

	return nil, errors.NewParserError("unexpected tag type: "+tp.tag.Kind, tp.tag.Start, tp.tag.End)
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

func (tp *TagParser) parseEndStatement() (Statement, error) {
	depth := len(tp.parser.unclosed)
	if depth == 0 {
		return nil, errors.NewParserError("unexpected end tag", tp.tag.Start, tp.tag.End)
	}

	switch tp.parser.unclosed[depth-1].Kind() {
	case "define":
		tp.parser.unclosed[depth-1].(*DefineStatement).EndTag = tp.tag
	case "slot":
		tp.parser.unclosed[depth-1].(*SlotStatement).EndTag = tp.tag
	case "for":
		tp.parser.unclosed[depth-1].(*ForStatement).EndTag = tp.tag
	}

	tp.parser.unclosed = tp.parser.unclosed[:depth-1]

	return &EndStatement{}, nil
}

func (tp *TagParser) parseExtendStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
	}

	return &ExtendStatement{
		Template: tp.token.Literal,
		tag:      tp.tag,
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
	if tp.token.Kind != tokenizer.TOK_DOT && tp.token.Kind != tokenizer.TOK_IDENT {
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
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseNumericStatement() (Statement, error) {
	val, err := strconv.Atoi(tp.token.Literal)
	if err != nil {
		return nil, errors.NewParserError("invalid integer", tp.tag.Start, tp.tag.End)
	}

	return &IntegerStatement{
		Value: val,
		tag:   tp.tag,
	}, nil
}

func (tp *TagParser) parseStringStatement() (Statement, error) {
	return &StringStatement{
		tag:   tp.tag,
		Value: tp.token.Literal,
	}, nil
}

func (tp *TagParser) parseVariableStatement() (*VariableStatement, error) {
	st := &VariableStatement{Parts: []Statement{}, tag: tp.tag}
	st.IsLocal = tp.token.Kind == tokenizer.TOK_DOT

	for tp.token != nil {
		if tp.token.Kind == tokenizer.TOK_DOT && !tp.expectNext(tokenizer.TOK_IDENT) {
			return nil, errors.NewParserError("misuse of dot notation, dot must be followed by an identifier", tp.tag.Start, tp.tag.End)
		}

		if tp.token.Kind == tokenizer.TOK_LPAREN {
			functionCall, err := tp.parseFunctionArgs()
			if err != nil {
				return nil, err
			}
			st.Parts = append(st.Parts, functionCall)
		} else if tp.token.Kind == tokenizer.TOK_DOT && tp.expectNext(tokenizer.TOK_IDENT) {
			tp.Next()
		} else if tp.token.Kind == tokenizer.TOK_IDENT {
			st.Parts = append(st.Parts, &VariablePartStatement{
				Name: tp.token.Literal,
				tag:  tp.tag,
			})
			tp.Next()
		} else {
			if tp.depth > 0 && tp.token.Kind == tokenizer.TOK_RPAREN {
				break
			}
			return nil, errors.NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start, tp.tag.End)
		}
	}

	return st, nil
}

func (tp *TagParser) parseFunctionArgs() (*FunctionCallStatement, error) {
	tp.Next()
	previousDepth := tp.depth
	tp.depth += 1
	fc := &FunctionCallStatement{Args: []Statement{}, tag: tp.tag}
	for tp.depth > previousDepth && tp.token != nil {
		if tp.token.Kind == tokenizer.TOK_DOT && !tp.expectNext(tokenizer.TOK_IDENT) {
			return nil, errors.NewParserError("misuse of dot notation, dot must be followed by an identifier", tp.tag.Start, tp.tag.End)
		}

		if tp.token.Kind == tokenizer.TOK_IDENT && tp.expectNext(tokenizer.TOK_LPAREN) {
			tp.Next()
			inner, err := tp.parseFunctionArgs()
			if err != nil {
				return nil, err
			}
			fc.Args = append(fc.Args, inner)

			continue
		} else if tp.token.Kind == tokenizer.TOK_RPAREN {
			tp.depth -= 1
		} else if tp.token.Kind == tokenizer.TOK_INTEGER {
			val, err := strconv.Atoi(tp.token.Literal)
			if err != nil {
				return nil, errors.NewParserError("invalid integer", tp.tag.Start, tp.tag.End)
			}
			fc.Args = append(fc.Args, &IntegerStatement{
				Value: val,
				tag:   tp.tag,
			})
		} else if tp.token.Kind == tokenizer.TOK_DOT && tp.expectNext(tokenizer.TOK_IDENT) {
			variableStatement, err := tp.parseVariableStatement()
			if err != nil {
				return nil, err
			}
			fc.Args = append(fc.Args, variableStatement)

			continue
		} else if tp.token.Kind == tokenizer.TOK_STRING {
			fc.Args = append(fc.Args, &StringStatement{
				Value: tp.token.Literal,
				tag:   tp.tag,
			})
		}

		tp.Next()
	}

	return fc, nil
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

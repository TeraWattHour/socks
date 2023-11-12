package parser

import (
	"github.com/terawatthour/socks/pkg/tokenizer"
	"strconv"
)

type TagProgram struct {
	Tag       *tokenizer.Tag
	Statement Statement
}

type Parser struct {
	Tokenizer *tokenizer.Tokenizer
	cursor    int
	Programs  []TagProgram
	tag       *tokenizer.Tag
	unclosed  []Statement
}

type TagParser struct {
	parser    *Parser
	tag       tokenizer.Tag
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

func NewTagParser(parser *Parser, tag tokenizer.Tag) *TagParser {
	return &TagParser{
		parser: parser,
		tag:    tag,
		cursor: -1,
	}
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

// Parse parses a tag, returns a statement (Statement) that can be evaluated and an error (*Error).
// This function is called individually for each tag.
func (tp *TagParser) Parse() (st Statement, err error) {
	tp.Next()

	if tp.token == nil {
		return nil, NewParserError("empty tag", tp.tag.Start)
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
			return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
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
			return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
		}
	} else if tp.tag.Kind == "execute" {
		switch tp.token.Kind {
		case tokenizer.TOK_FOR:
			return tp.parseForStatement()
		case tokenizer.TOK_END:
			return tp.parseEndStatement()
		}
	}

	return nil, NewParserError("unexpected tag type: "+tp.tag.Kind, tp.tag.Start)
}

func (tp *TagParser) parseEndStatement() (Statement, error) {
	depth := len(tp.parser.unclosed)
	if depth == 0 {
		return nil, NewParserError("unexpected end tag", tp.tag.Start)
	}

	switch tp.parser.unclosed[depth-1].Kind() {
	case "define":
		tp.parser.unclosed[depth-1].(*DefineStatement).EndTag = &tp.tag
	case "slot":
		tp.parser.unclosed[depth-1].(*SlotStatement).EndTag = &tp.tag
	case "for":
		tp.parser.unclosed[depth-1].(*ForStatement).EndTag = &tp.tag
	}

	tp.parser.unclosed = tp.parser.unclosed[:depth-1]

	return &EndStatement{}, nil
}

func (tp *TagParser) parseExtendStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}

	return &ExtendStatement{
		Template: tp.token.Literal,
	}, nil
}

func (tp *TagParser) parseDefineStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}

	statement := &DefineStatement{
		Name:     tp.token.Literal,
		StartTag: &tp.tag,
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
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}
	iteratorName = tp.token.Literal
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_COMMA {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IDENT {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}
	valueName = tp.token.Literal
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_IN {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_DOT && tp.token.Kind != tokenizer.TOK_IDENT {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}
	iterable, err := tp.parseVariableStatement()
	if err != nil {
		return nil, err
	}

	statement := &ForStatement{
		IteratorName: iteratorName,
		ValueName:    valueName,
		Iterable:     iterable,
		StartTag:     &tp.tag,
		EndTag:       nil,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseSlotStatement() (Statement, error) {
	tp.Next()
	if tp.token.Kind != tokenizer.TOK_STRING {
		return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
	}

	statement := &SlotStatement{
		Name:     tp.token.Literal,
		StartTag: &tp.tag,
	}

	tp.parser.unclosed = append(tp.parser.unclosed, statement)

	return statement, nil
}

func (tp *TagParser) parseNumericStatement() (Statement, error) {
	val, err := strconv.Atoi(tp.token.Literal)
	if err != nil {
		return nil, NewParserError("invalid integer", tp.tag.Start)
	}

	return &IntegerStatement{
		Value: val,
	}, nil
}

func (tp *TagParser) parseStringStatement() (Statement, error) {
	return &StringStatement{
		Value: tp.token.Literal,
	}, nil
}

func (tp *TagParser) parseVariableStatement() (*VariableStatement, error) {
	st := &VariableStatement{Parts: []Statement{}}
	st.IsLocal = tp.token.Kind == tokenizer.TOK_DOT

	for tp.token != nil {
		if tp.token.Kind == tokenizer.TOK_DOT && !tp.expectNext(tokenizer.TOK_IDENT) {
			return nil, NewParserError("misuse of dot notation, dot must be followed by an identifier", tp.tag.Start)
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
			})
			tp.Next()
		} else {
			if tp.depth > 0 && tp.token.Kind == tokenizer.TOK_RPAREN {
				break
			}
			return nil, NewParserError("unexpected token: "+tp.token.Literal, tp.tag.Start)
		}
	}

	return st, nil
}

func (tp *TagParser) parseFunctionArgs() (*FunctionCallStatement, error) {
	tp.Next()
	previousDepth := tp.depth
	tp.depth += 1
	fc := &FunctionCallStatement{Args: []Statement{}}
	for tp.depth > previousDepth && tp.token != nil {
		if tp.token.Kind == tokenizer.TOK_DOT && !tp.expectNext(tokenizer.TOK_IDENT) {
			return nil, NewParserError("misuse of dot notation, dot must be followed by an identifier", tp.tag.Start)
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
				return nil, NewParserError("invalid integer", tp.tag.Start)
			}
			fc.Args = append(fc.Args, &IntegerStatement{
				Value: val,
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

func (p *Parser) Parse() error {
	p.Next()

	var programs []TagProgram

	for p.tag != nil {
		statement, err := NewTagParser(p, *p.tag).Parse()
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
		panic("unclosed tag")
	}

	p.Programs = programs
	return nil
}

func (p *Parser) Next() {
	p.cursor += 1

	if p.cursor >= len(p.Tokenizer.Tags) {
		p.tag = nil
	} else {
		p.tag = &p.Tokenizer.Tags[p.cursor]
	}
}

package tokenizer

import (
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"unicode"
)

const (
	TOK_DOT     = "dot"
	TOK_LPAREN  = "lparen"
	TOK_RPAREN  = "rparen"
	TOK_IDENT   = "ident"
	TOK_STRING  = "string"
	TOK_COMMA   = "comma"
	TOK_INTEGER = "integer"
	TOK_FLOAT   = "float"
	TOK_FOR     = "for"
	TOK_IN      = "in"
	TOK_EXTEND  = "extend"
	TOK_SLOT    = "slot"
	TOK_END     = "end"
	TOK_DEFINE  = "define"
)

type Tokenizer struct {
	Tags          []Tag
	Template      string
	Runes         []rune
	currentTag    *Tag
	cursor        int
	char          rune
	nextChar      rune
	isInsideBlock bool
	isInsideQuote bool
}

type Tag struct {
	Start int
	End   int
	// Tag.Kind is either "print", "preprocessor" or "execute"
	Kind   string
	Tokens []Token
	Body   string
}

type Token struct {
	Kind    string
	Literal string
}

var KEYWORDS = []string{
	TOK_FOR,
	TOK_IN,
	TOK_EXTEND,
	TOK_SLOT,
	TOK_END,
	TOK_DEFINE,
}

func NewTokenizer(template string) *Tokenizer {
	t := &Tokenizer{
		Template: template,
		Runes:    []rune(template),
		cursor:   -1,
	}

	t.Next()

	return t
}

func (t *Tokenizer) Tokenize() error {
	for t.char != 0 {
		pushNext := true

		t.skipWhitespace()

		if t.isInsideBlock {
			var token Token

			switch t.char {
			case '.':
				token = Token{
					Kind:    TOK_DOT,
					Literal: ".",
				}
			case '(':
				token = Token{
					Kind:    TOK_LPAREN,
					Literal: "(",
				}
			case '"':
				t.Next()
				start := t.cursor
				for t.char != '"' {
					t.Next()
				}
				literal := string(t.Runes[start:t.cursor])
				t.Next()
				token = Token{
					Kind:    TOK_STRING,
					Literal: literal,
				}
				pushNext = false
			case ',':
				token = Token{
					Kind:    TOK_COMMA,
					Literal: ",",
				}
			case ')':
				token = Token{
					Kind:    TOK_RPAREN,
					Literal: ")",
				}
			default:
				if t.isValidNumber() {
					start := t.cursor
					hasDot := t.char == '.'
					for t.isValidNumber() || t.char == '.' {
						t.Next()
						isDot := t.char == '.'
						if hasDot && isDot {
							return errors2.NewTokenizerError("malformed number", start, t.cursor)
						}
						if isDot {
							hasDot = true
						}
					}

					if hasDot {
						token = Token{
							Kind:    TOK_FLOAT,
							Literal: string(t.Runes[start:t.cursor]),
						}
					} else {
						token = Token{
							Kind:    TOK_INTEGER,
							Literal: string(t.Runes[start:t.cursor]),
						}
					}

					pushNext = false
				} else if t.isValidVariableName() {
					start := t.cursor
					for t.isValidVariableName() || t.isValidNumber() {
						t.Next()
					}
					literal := string(t.Runes[start:t.cursor])
					if helpers.Contains(KEYWORDS, literal) {
						token = Token{
							Kind:    literal,
							Literal: literal,
						}
					} else {
						token = Token{
							Kind:    TOK_IDENT,
							Literal: string(t.Runes[start:t.cursor]),
						}
					}
					pushNext = false
				}
			}

			if token.Kind != "" {
				t.currentTag.Tokens = append(t.currentTag.Tokens, token)
			}

			if err := t.tryCloseTag(); err != nil {
				return err
			}
		} else {
			t.tryOpenTag()
		}

		if pushNext {
			t.Next()
		}
	}

	if t.isInsideBlock {
		return errors2.NewTokenizerError("unexpected end of file", t.cursor, t.cursor)
	}

	return nil
}

func (t *Tokenizer) Next() {
	t.cursor += 1
	if t.cursor >= len(t.Runes) {
		t.char = 0
	} else {
		t.char = t.Runes[t.cursor]
	}

	if t.cursor+1 >= len(t.Runes) {
		t.nextChar = 0
	} else {
		t.nextChar = t.Runes[t.cursor+1]
	}
}

func (t *Tokenizer) tryOpenTag() {
	openTag := func(kind string) {
		t.currentTag = &Tag{
			Start: t.cursor,
			Kind:  kind,
		}
		t.isInsideBlock = true
	}

	if t.char == '{' {
		switch t.nextChar {
		case '%':
			openTag("preprocessor")
		case '!':
			openTag("execute")
		case '{':
			openTag("print")
		}
	}
}

func (t *Tokenizer) tryCloseTag() error {
	closeTag := func(kind string) error {
		if t.currentTag == nil || t.currentTag.Kind != kind {
			return errors2.NewTokenizerError("unexpected tag terminator", t.cursor, t.cursor)
		}
		t.currentTag.End = t.cursor + 1

		t.Tags = append(t.Tags, *t.currentTag)
		t.currentTag = nil

		t.isInsideBlock = false

		return nil
	}

	if t.nextChar == '}' {
		switch t.char {
		case '}':
			return closeTag("print")
		case '%':
			return closeTag("preprocessor")
		case '!':
			return closeTag("execute")
		}
	}

	return nil
}

func (t *Tokenizer) skipWhitespace() {
	for t.char == ' ' || t.char == '\t' || t.char == '\n' || t.char == '\r' {
		t.Next()
	}
}

func (t *Tokenizer) isValidVariableName() bool {
	return unicode.IsLetter(t.char) || t.char == '_'
}

func (t *Tokenizer) isValidNumber() bool {
	return t.char >= '0' && t.char <= '9'
}

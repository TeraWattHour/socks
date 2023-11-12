package tokenizer

import (
	"errors"
	"golang.org/x/exp/slices"
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
	Template      string
	cursor        int
	Tags          []Tag
	Runes         []rune
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

func (t *Tokenizer) Tokenize() error {
	var currentTag *Tag

	for t.char != 0 {
		pushNext := true

		t.skipWhitespace()

		if t.isInsideBlock {
			var token Token

			// inside print block
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
					for t.isValidNumber() {
						t.Next()
						isDot := t.char == '.'
						if hasDot && isDot {
							return errors.New("malformed number")
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
					for t.isValidVariableName() {
						t.Next()
					}
					literal := string(t.Runes[start:t.cursor])
					if slices.Contains(KEYWORDS, literal) {
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
				currentTag.Tokens = append(currentTag.Tokens, token)
			}

			if t.char == '}' && t.nextChar == '}' {
				if currentTag == nil || currentTag.Kind != "print" {
					panic("unexpected }}")
				}
				currentTag.End = t.cursor + 1

				t.Tags = append(t.Tags, *currentTag)
				currentTag = nil

				t.isInsideBlock = false
			}

			if t.char == '%' && t.nextChar == '}' {
				if currentTag == nil || currentTag.Kind != "preprocessor" {
					panic("unexpected %}")
				}

				currentTag.End = t.cursor + 1

				t.Tags = append(t.Tags, *currentTag)
				currentTag = nil

				t.isInsideBlock = false
			}

			if t.char == '!' && t.nextChar == '}' {
				if currentTag == nil || currentTag.Kind != "execute" {
					panic("unexpected %}")
				}

				currentTag.End = t.cursor + 1

				t.Tags = append(t.Tags, *currentTag)
				currentTag = nil

				t.isInsideBlock = false
			}

		} else {
			if t.char == '{' && t.nextChar == '!' {
				currentTag = &Tag{
					Start: t.cursor,
					Kind:  "execute",
				}
				t.isInsideBlock = true
			}

			if t.char == '{' && t.nextChar == '{' {
				currentTag = &Tag{
					Start: t.cursor,
					Kind:  "print",
				}
				t.isInsideBlock = true
			}

			if t.char == '{' && t.nextChar == '%' {
				currentTag = &Tag{
					Start: t.cursor,
					Kind:  "preprocessor",
				}
				t.isInsideBlock = true
			}
		}

		if pushNext {
			t.Next()
		}
	}

	if t.isInsideBlock {
		return errors.New("unexpected end of file")
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

func (t *Tokenizer) skipWhitespace() {
	for t.char == ' ' || t.char == '\t' || t.char == '\n' || t.char == '\r' {
		t.Next()
	}
}

func (t *Tokenizer) isValidVariableName() bool {
	if t.char >= 'a' && t.char <= 'z' || t.char >= 'A' && t.char <= 'Z' || t.char == '_' {
		return true
	}
	return false
}

func (t *Tokenizer) isValidNumber() bool {
	if t.char >= '0' && t.char <= '9' || t.char == '.' {
		return true
	}
	return false
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

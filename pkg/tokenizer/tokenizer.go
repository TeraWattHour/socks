package tokenizer

import (
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"unicode"
)

const (
	TOK_UNKNOWN = "unknown" // TOK_UNKNOWN is used for tokens that needn't be recognised by the templating engine but may be used in evaluation

	TOK_IDENT  = "ident"
	TOK_STRING = "string"
	TOK_COMMA  = "comma"
	TOK_END    = "end"

	// execute keywords
	TOK_FOR = "for"
	TOK_IN  = "in"
	TOK_IF  = "if"

	// preprocessor keywords
	TOK_EXTEND   = "extend"
	TOK_SLOT     = "slot"
	TOK_TEMPLATE = "template"
	TOK_DEFINE   = "define"
)

var KEYWORDS = []string{
	TOK_FOR,
	TOK_IN,
	TOK_EXTEND,
	TOK_SLOT,
	TOK_END,
	TOK_DEFINE,
	TOK_TEMPLATE,
	TOK_IF,
}

type Tokenizer struct {
	Tags          []Tag
	Template      string
	Runes         []rune
	currentTag    *Tag
	cursor        int
	char          rune
	nextChar      rune
	isInsideTag   bool
	isInsideQuote bool
}

type TagKind string

const (
	PrintKind        TagKind = "print"
	PreprocessorKind TagKind = "preprocessor"
	ExecuteKind      TagKind = "execute"
)

type Tag struct {
	Start  int
	End    int
	Kind   TagKind // Tag.Kind is either "print", "preprocessor" or "execute"
	Tokens []Token
	Body   string
}

type Token struct {
	Kind    string
	Literal string
	Start   int
	Length  int
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

		if t.isInsideTag {
			token := Token{Start: t.cursor, Length: 1}

			switch t.char {
			case '"', '\'':
				quoteChar := t.char
				t.Next()
				start := t.cursor
				for t.char != quoteChar {
					t.Next()
				}
				literal := string(t.Runes[start:t.cursor])
				t.Next()
				token.Kind = TOK_STRING
				token.Literal = literal
				token.Start = start - 1
				token.Length = t.cursor - start + 1
				pushNext = false
			case ',':
				token.Kind = TOK_COMMA
				token.Literal = ","
			default:
				if t.isValidVariableName() {
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
					token.Start = start
					token.Length = t.cursor - start
					pushNext = false
				} else {
					token.Kind = TOK_UNKNOWN
					token.Literal = string(t.char)
				}
			}

			if t.currentTag != nil && token.Kind != "" && !(pushNext && t.canCloseTag()) {
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

	if t.isInsideTag {
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
	openTag := func(kind TagKind) {
		t.currentTag = &Tag{
			Start: t.cursor,
			Kind:  kind,
		}
		t.isInsideTag = true
		t.Next()
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

func (t *Tokenizer) canCloseTag() bool {
	return t.nextChar == '}' && (t.char == '}' || t.char == '%' || t.char == '!')
}

func (t *Tokenizer) tryCloseTag() error {
	closeTag := func(kind TagKind) error {
		if t.currentTag == nil || t.currentTag.Kind != kind {
			return errors2.NewTokenizerError("unexpected tag terminator", t.cursor, t.cursor)
		}
		t.currentTag.End = t.cursor + 1

		t.Tags = append(t.Tags, *t.currentTag)
		t.currentTag = nil

		t.isInsideTag = false

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

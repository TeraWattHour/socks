package tokenizer

import (
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"regexp"
	"unicode"
)

var OpenTagRe = regexp.MustCompile(`(\\)?({{?|{#|{!|{%})`)

const (
	TokUnknown = "unknown" // TokUnknown is used for tokens that needn't be recognised by the templating engine but may be used in evaluation

	TokIdent  = "ident"
	TokString = "string"
	TokComma  = "comma"
	TokEnd    = "end"

	TokFor = "for"
	TokIn  = "in"
	TokIf  = "if"

	TokExtend   = "extend"
	TokSlot     = "slot"
	TokTemplate = "template"
	TokDefine   = "define"
)

var KEYWORDS = []string{
	TokFor,
	TokIn,
	TokExtend,
	TokSlot,
	TokEnd,
	TokDefine,
	TokTemplate,
	TokIf,
}

type Tokenizer struct {
	Template string
	Runes    []rune
	Tags     []Tag

	currentTag  *Tag
	cursor      int
	char        rune
	nextChar    rune
	isInsideTag bool
}

// TagKind is either PrintKind, PreprocessorKind, StaticKind, or ExecuteKind.
type TagKind string

const (
	PrintKind        TagKind = "print"
	PreprocessorKind TagKind = "preprocessor"
	StaticKind       TagKind = "static"
	ExecuteKind      TagKind = "execute"
	CommentKind      TagKind = "comment"
)

type Tag struct {
	Start  int
	End    int
	Kind   TagKind
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
		Tags:     make([]Tag, 0),
		cursor:   -1,
	}

	t.Next()

	return t
}

func findTagOpenings(template string) []int {
	return helpers.Map(OpenTagRe.FindAllStringIndex(template, -1), func(loc []int) int {
		return loc[0]
	})
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
				token.Kind = TokString
				token.Literal = literal
				token.Start = start - 1
				token.Length = t.cursor - start + 1
				pushNext = false
			case ',':
				token.Kind = TokComma
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
							Kind:    TokIdent,
							Literal: string(t.Runes[start:t.cursor]),
						}
					}
					token.Start = start
					token.Length = t.cursor - start
					pushNext = false
				} else {
					token.Kind = TokUnknown
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
		t.Next()
	}

	if t.char == '{' {
		switch t.nextChar {
		case '%':
			openTag(PreprocessorKind)
		case '!':
			openTag(ExecuteKind)
		case '{':
			openTag(PrintKind)
		case '#':
			openTag(CommentKind)
		case '$':
			openTag(StaticKind)
		}
	}
}

func (t *Tokenizer) canCloseTag() bool {
	return t.nextChar == '}' && (t.char == '}' || t.char == '%' || t.char == '!' || t.char == '$')
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
			return closeTag(PrintKind)
		case '%':
			return closeTag(PreprocessorKind)
		case '!':
			return closeTag(ExecuteKind)
		case '#':
			return closeTag(CommentKind)
		case '$':
			return closeTag(StaticKind)
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

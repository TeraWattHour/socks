package tokenizer

import (
	"fmt"
	"slices"
	"unicode"

	errors2 "github.com/terawatthour/socks/pkg/errors"
)

var StatementKeywords = []string{
	TokFor,
	TokIf,
	"define",
	"extend",
	"slot",
	"template",
	"endif",
	"endfor",
	"enddefine",
	"endtemplate",
	"endslot",
}

const (
	TokUnknown = "unknown" // TokUnknown is used for tokens that needn't be recognised by the templating engine but may be used in evaluation

	TokIdent  = "ident"
	TokNumber = "number"
	TokString = "string"
	TokComma  = "comma"
	TokEnd    = "end"
	TokAt     = "at"

	TokLparen = "lparen"
	TokRparen = "rparen"
	TokLbrack = "lbrack"
	TokRbrack = "rbrack"
	TokLbrace = "lbrace"
	TokRbrace = "rbrace"

	TokLt  = "lt"
	TokGt  = "gt"
	TokEq  = "eq"
	TokNeq = "neq"
	TokLte = "lte"
	TokGte = "gte"

	TokAmpersand     = "ampersand"
	TokBang          = "bang"
	TokPlus          = "plus"
	TokMinus         = "minus"
	TokAsterisk      = "asterisk"
	TokSlash         = "slash"
	TokModulo        = "modulo"
	TokPower         = "power"
	TokFloorDiv      = "floor_div"
	TokColon         = "colon"
	TokQuestion      = "question"
	TokDot           = "dot"
	TokOptionalChain = "optional_chain"

	TokFor   = "for"
	TokIn    = "in"
	TokIf    = "if"
	TokTrue  = "true"
	TokNot   = "not"
	TokFalse = "false"
	TokAnd   = "and"
	TokOr    = "or"

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
	TokTrue,
	TokFalse,
	TokAnd,
	TokOr,
	TokNot,
}

type Element interface {
	Kind() string
	String() string
	Tokens() []Token
}

type Text string

func (t Text) Kind() string {
	return "text"
}

func (t Text) String() string {
	return string(t)
}

func (t Text) Tokens() []Token {
	return nil
}

type Tokenizer struct {
	Template []rune
	Elements []Element

	currentTag  *Tag
	cursor      int
	prevChar    rune
	char        rune
	nextChar    rune
	isInsideTag bool
	lastClosing int
	line        int
	column      int
}

// TagType is either PrintKind, PreprocessorKind, StaticKind, or ExecuteKind.
type TagType string

const (
	PrintKind   TagType = "print"
	CommentKind TagType = "comment"
)

type Tag struct {
	Start    int
	End      int
	Literal  string
	Sanitize bool
	Type     TagType
	tokens   []Token
}

func (t *Tag) Tokens() []Token {
	return t.tokens
}

func (t *Tag) Kind() string {
	return "tag"
}

func (t *Tag) String() string {
	return t.Literal
}

type Statement struct {
	Instruction string
	tokens      []Token
	Flags       []string
}

func (s *Statement) Tokens() []Token {
	return s.tokens
}

func (s *Statement) Kind() string {
	return "statement"
}

func (s *Statement) String() string {
	return s.Instruction
}

type Token struct {
	Kind     string
	Literal  string
	Start    int
	Length   int
	Location Location
}

type Location struct {
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("[type: %s, literal: %s]", t.Kind, t.Literal)
}

func NewTokenizer(template string) *Tokenizer {
	t := &Tokenizer{
		Template: []rune(template),
		Elements: make([]Element, 0),
		cursor:   -1,
		line:     1,
		column:   0,
	}

	t.Next()

	return t
}

func (t *Tokenizer) Tokenize() error {
	for t.char != 0 {
		t.skipWhitespace()

		if t.isInsideTag {
			t.tokenizeExpression()
			continue
		} else {
			t.tryOpenTag()
			if t.isInsideTag {
				continue
			}
		}

		if t.char == '@' && t.prevChar != '\\' {
			t.grabText(t.cursor)
			t.Next()
			if t.isAsciiLetter() {
				start := t.cursor
				for t.isAsciiLetter() {
					t.Next()
				}
				literal := string(t.Template[start:t.cursor])
				if !isValidStatementLiteral(literal) {
					return errors2.NewError(fmt.Sprintf("unexpected instruction '%s'", literal))
				}
				statement := Statement{
					Instruction: literal,
					tokens:      make([]Token, 0),
				}
				t.skipWhitespace()
				if t.char == '[' {
					t.Next()
					t.skipWhitespace()
					if t.char == ']' {
						return errors2.NewError("unexpected empty flag set")
					}
					flags := make([]string, 0)
					for t.char != ']' {
						start := t.cursor
						for t.isAsciiLetter() {
							t.Next()
						}
						flags = append(flags, string(t.Template[start:t.cursor]))
						t.skipWhitespace()
						if t.char == ']' {
							break
						}
						if t.char != ',' {
							return errors2.NewError(fmt.Sprintf("unexpected token in flag set: '%s'", string(t.char)))
						}
						t.Next()
					}
					statement.Flags = flags
					t.Next()
					t.skipWhitespace()
				}
				if t.char != '(' {
					t.Elements = append(t.Elements, &statement)
					t.lastClosing = t.cursor
					continue
				}
				tokens := t.tokenizeExpression()
				statement.tokens = append(statement.tokens, tokens...)
				t.Elements = append(t.Elements, &statement)
			}
		}

		t.Next()
	}

	if t.isInsideTag {
		return errors2.NewError("unexpected end of template")
	}

	t.Elements = append(t.Elements, Text(t.Template[t.lastClosing:len(t.Template)]))

	return nil
}

func (t *Tokenizer) tokenizeExpression() []Token {
	depth := 0

	tokens := make([]Token, 0)

	for t.char != 0 {
		t.skipWhitespace()

		if t.isInsideTag {
			if closed, err := t.tryCloseTag(); err != nil {
				panic(err)
			} else if closed {
				return tokens
			}
		}

		pushNext := true
		token := Token{Start: t.cursor, Length: 1, Literal: string(t.char), Location: Location{t.line, t.column}}

		switch t.char {
		case '.':
			token.Kind = TokDot
		case '&':
			token.Kind = TokAmpersand
		case '?':
			if t.nextChar == '.' {
				token.Kind = TokOptionalChain
				token.Literal = "?."
				t.Next()
			} else {
				token.Kind = TokQuestion
			}
		case ':':
			token.Kind = TokColon
		case '%':
			token.Kind = TokModulo
		case '[':
			token.Kind = TokLbrack
			depth += 1
		case ']':
			token.Kind = TokRbrack
			depth -= 1
		case '{':
			token.Kind = TokLbrace
			depth += 1
		case '}':
			token.Kind = TokRbrace
			depth -= 1
		case '@':
			token.Kind = TokAt
		case '+':
			token.Kind = TokPlus
		case '-':
			token.Kind = TokMinus
		case '*':
			if t.nextChar == '*' {
				token.Kind = TokPower
				token.Literal = "**"
				t.Next()
			} else {
				token.Kind = TokAsterisk
			}
		case '/':
			if t.nextChar == '/' {
				token.Kind = TokFloorDiv
				token.Literal = "//"
				t.Next()
			} else {
				token.Kind = TokSlash
			}
		case '<':
			if t.nextChar == '=' {
				token.Kind = TokLte
				token.Literal = "<="
				t.Next()
			} else {
				token.Kind = TokLt
			}
		case '>':
			if t.nextChar == '=' {
				token.Kind = TokGte
				token.Literal = ">="
				t.Next()
			} else {
				token.Kind = TokGt
			}
		case '!':
			if t.nextChar == '=' {
				token.Kind = TokNeq
				token.Literal = "!="
				t.Next()
			} else {
				token.Kind = TokBang
			}
		case '=':
			if t.nextChar == '=' {
				token.Kind = TokEq
				token.Literal = "=="
				t.Next()
			} else {
				panic("unexpected token '='")
			}
		case '"', '\'':
			quoteChar := t.char
			t.Next()
			start := t.cursor
			for t.char != quoteChar {
				t.Next()
			}
			literal := string(t.Template[start:t.cursor])
			t.Next()
			token.Kind = TokString
			token.Literal = literal
			token.Start = start - 1
			token.Length = t.cursor - start + 1
			pushNext = false
		case ',':
			token.Kind = TokComma
		case '(':
			token.Kind = TokLparen
			depth += 1
		case ')':
			token.Kind = TokRparen
			depth -= 1
		default:
			if t.isValidVariableName() {
				start := t.cursor
				for t.isValidVariableName() || t.isDigit() {
					t.Next()
				}
				literal := string(t.Template[start:t.cursor])
				if slices.Index(KEYWORDS, literal) != -1 {
					token = Token{
						Kind:    literal,
						Literal: literal,
					}
				} else {
					token = Token{
						Kind:    TokIdent,
						Literal: string(t.Template[start:t.cursor]),
					}
				}
				token.Start = start
				token.Length = t.cursor - start
				pushNext = false
			} else if t.isDigit() {
				start := t.cursor
				hasDot := false
				for t.isDigit() || t.char == '.' {
					t.Next()
					if t.char == '.' {
						if hasDot {
							panic("unexpected dot in number")
						}
						hasDot = true
					}
				}
				literal := string(t.Template[start:t.cursor])
				token = Token{
					Kind:    TokNumber,
					Literal: literal,
					Start:   start,
					Length:  t.cursor - start,
				}
				pushNext = false
			} else {
				token.Kind = TokUnknown
				token.Literal = string(t.char)
			}
		}

		tokens = append(tokens, token)
		if t.isInsideTag {
			t.currentTag.tokens = append(t.currentTag.tokens, token)
		}
		if !t.isInsideTag && depth == 0 {
			t.lastClosing = t.cursor + 1
			break
		}

		if pushNext {
			t.Next()
		}
	}

	return tokens
}

func (t *Tokenizer) Next() {
	t.cursor += 1
	t.prevChar = t.char
	if t.cursor >= len(t.Template) {
		t.char = 0
	} else {
		t.char = t.Template[t.cursor]
	}

	if t.cursor+1 >= len(t.Template) {
		t.nextChar = 0
	} else {
		t.nextChar = t.Template[t.cursor+1]
	}

	if t.char == '\n' {
		t.line += 1
		t.column = 0
	} else {
		t.column += 1
	}
}

func (t *Tokenizer) grabText(cursor int) {
	if t.lastClosing < cursor && cursor < len(t.Template) {
		t.Elements = append(t.Elements, Text(t.Template[t.lastClosing:cursor]))
	}
}

func (t *Tokenizer) tryOpenTag() {
	if t.prevChar == '\\' {
		return
	}

	openTag := func(kind TagType) {
		t.grabText(t.cursor)

		t.currentTag = &Tag{
			Start: t.cursor,
			Type:  kind,
		}
		t.Next()
		t.currentTag.Sanitize = t.char == '{'
		t.isInsideTag = true
		t.Next()
	}

	if t.char == '{' {
		switch t.nextChar {
		case '{':
			openTag(PrintKind)
		case '#':
			openTag(CommentKind)
		case '!':
			openTag(PrintKind)
		}
	}
}

func (t *Tokenizer) tryCloseTag() (bool, error) {
	closeTag := func(kind TagType) (bool, error) {
		if t.currentTag == nil || t.currentTag.Type != kind || t.currentTag.Sanitize != (t.char == '}') {
			return false, errors2.NewError("unexpected tag termination")
		}
		t.currentTag.End = t.cursor + 2
		t.currentTag.Literal = string(t.Template[t.currentTag.Start:t.currentTag.End])
		t.lastClosing = t.cursor + 2
		t.Elements = append(t.Elements, t.currentTag)
		t.currentTag = nil

		t.Next()
		t.Next()

		t.isInsideTag = false

		return true, nil
	}

	if t.nextChar == '}' {
		switch t.char {
		case '}':
			return closeTag(PrintKind)
		case '#':
			return closeTag(CommentKind)
		case '!':
			return closeTag(PrintKind)
		}
	}

	return false, nil
}

func (t *Tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.char) {
		t.Next()
	}
}

func (t *Tokenizer) isValidVariableName() bool {
	return unicode.IsLetter(t.char) || t.char == '_'
}

func (t *Tokenizer) isDigit() bool {
	return t.char >= '0' && t.char <= '9'
}

func (t *Tokenizer) isAsciiLetter() bool {
	return t.char >= 'a' && t.char <= 'z' || t.char >= 'A' && t.char <= 'Z' || t.char == '_'
}

func isValidStatementLiteral(literal string) bool {
	return slices.Index(StatementKeywords, literal) != -1
}

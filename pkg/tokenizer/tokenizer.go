package tokenizer

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"slices"
	"unicode"

	errors2 "github.com/terawatthour/socks/pkg/errors"
)

type _tokenizer struct {
	template []rune
	elements []Element

	currentTag  *Mustache
	cursor      int
	prevChar    rune
	char        rune
	nextChar    rune
	isInsideTag bool
	lastClosing int
	line        int
	column      int
	location    helpers.Location
}

type Token struct {
	Kind          string
	Literal       string
	Start         int
	Length        int
	LocationStart helpers.Location
	LocationEnd   helpers.Location
}

func (t Token) String() string {
	return fmt.Sprintf("[type: %s, literal: %s]", t.Kind, t.Literal)
}

func Tokenize(template string) ([]Element, error) {
	t := &_tokenizer{
		template: []rune(template),
		elements: make([]Element, 0),
		cursor:   -1,
		line:     1,
		column:   0,
	}

	t.next()

	return t.tokenize()
}

func (t *_tokenizer) tokenize() ([]Element, error) {
	for t.char != 0 {
		t.skipWhitespace()

		if t.tryOpenMustache() {
			t.tokenizeExpression()
			continue
		}

		// open an instruction tag
		if t.char == '@' && t.prevChar != '\\' {
			loc := helpers.Location{t.line, t.column}
			t.next()
			if isAsciiLetter(t.nextChar) {
				start := t.cursor
				for isAsciiLetter(t.char) {
					t.next()
				}
				endOfInstruction := t.cursor
				literal := string(t.template[start:endOfInstruction])
				if !isValidStatementLiteral(literal) {
					continue
				}
				t.grabText(start - 1)

				statement := Statement{
					Instruction: literal,
					Tokens:      make([]Token, 0),
					Location:    loc,
				}
				t.skipWhitespace()
				if t.char == '[' {
					t.next()
					t.skipWhitespace()
					if t.char == ']' {
						return nil, errors2.NewErrorWithLocation("unexpected empty flag set", t.location)
					}
					flags := make([]string, 0)
					for t.char != ']' {
						start := t.cursor
						for isAsciiLetter(t.char) {
							t.next()
						}
						flags = append(flags, string(t.template[start:t.cursor]))
						t.skipWhitespace()
						if t.char == ']' {
							break
						}
						if t.char != ',' {
							return nil, errors2.NewError(fmt.Sprintf("unexpected token in flag set: '%s'", string(t.char)))
						}
						t.next()
					}
					statement.Flags = flags
					t.next()
					t.skipWhitespace()
				}
				if t.char != '(' {
					t.elements = append(t.elements, &statement)
					t.lastClosing = endOfInstruction
					continue
				}
				tokens := t.tokenizeExpression()
				statement.Tokens = append(statement.Tokens, tokens...)
				t.elements = append(t.elements, &statement)
				continue
			}
		}

		t.next()
	}

	if t.isInsideTag {
		return nil, errors2.NewError("unexpected end of template")
	}

	t.elements = append(t.elements, Text(t.template[t.lastClosing:]))

	return t.elements, nil
}

func (t *_tokenizer) tokenizeExpression() []Token {
	depth := 0

	tokens := make([]Token, 0)

	for t.char != 0 {
		t.skipWhitespace()

		if t.isInsideTag {
			if closed, err := t.tryCloseMustache(); err != nil {
				panic(err)
			} else if closed {
				return tokens
			}
		}

		pushNext := true
		token := Token{Start: t.cursor, Length: 1, Literal: string(t.char), LocationStart: helpers.Location{t.line, t.column}, LocationEnd: helpers.Location{t.line, t.column + 1}}

		switch t.char {
		case '.':
			token.Kind = TokDot
		case '&':
			token.Kind = TokAmpersand
		case '?':
			if t.nextChar == '.' {
				token.Kind = TokOptionalChain
				token.Literal = "?."
				t.next()
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
				t.next()
			} else {
				token.Kind = TokAsterisk
			}
		case '/':
			if t.nextChar == '/' {
				token.Kind = TokFloorDiv
				token.Literal = "//"
				t.next()
			} else {
				token.Kind = TokSlash
			}
		case '<':
			if t.nextChar == '=' {
				token.Kind = TokLte
				token.Literal = "<="
				t.next()
			} else {
				token.Kind = TokLt
			}
		case '>':
			if t.nextChar == '=' {
				token.Kind = TokGte
				token.Literal = ">="
				t.next()
			} else {
				token.Kind = TokGt
			}
		case '!':
			if t.nextChar == '=' {
				token.Kind = TokNeq
				token.Literal = "!="
				t.next()
			} else {
				token.Kind = TokBang
			}
		case '=':
			if t.nextChar == '=' {
				token.Kind = TokEq
				token.Literal = "=="
				t.next()
			} else {
				panic("unexpected token '='")
			}
		case '"', '\'':
			quoteChar := t.char
			t.next()
			start := t.cursor
			for t.char != quoteChar {
				if t.char == 0 {
					panic("unterminated string literal")
				}
				t.next()
			}
			literal := string(t.template[start:t.cursor])
			t.next()
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
					t.next()
				}
				literal := string(t.template[start:t.cursor])
				if slices.Index(Keywords, literal) != -1 {
					token = Token{
						Kind:    literal,
						Literal: literal,
					}
				} else {
					token = Token{
						Kind:    TokIdent,
						Literal: string(t.template[start:t.cursor]),
					}
				}
				token.Start = start
				token.Length = t.cursor - start
				pushNext = false
			} else if t.isDigit() {
				start := t.cursor
				hasDot := false
				for t.isDigit() || t.char == '.' {
					t.next()
					if t.char == '.' {
						if hasDot {
							panic("unexpected dot in number")
						}
						hasDot = true
					}
				}
				literal := string(t.template[start:t.cursor])
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

		token.Length = t.cursor - token.Start
		if pushNext {
			t.next()
		}

		token.LocationEnd = helpers.Location{t.line, t.column}

		tokens = append(tokens, token)
		if t.isInsideTag {
			t.currentTag.Tokens = append(t.currentTag.Tokens, token)
		}
		if !t.isInsideTag && depth == 0 {
			t.lastClosing = t.cursor
			break
		}

	}

	return tokens
}

func (t *_tokenizer) next() {
	t.cursor += 1
	t.prevChar = t.char
	if t.cursor >= len(t.template) {
		t.char = 0
	} else {
		t.char = t.template[t.cursor]
	}

	if t.cursor+1 >= len(t.template) {
		t.nextChar = 0
	} else {
		t.nextChar = t.template[t.cursor+1]
	}

	if t.char == '\n' {
		t.line += 1
		t.column = 0
	} else {
		t.column += 1
	}

	t.location = helpers.Location{t.line, t.column}
}

func (t *_tokenizer) grabText(cursor int) {
	if t.lastClosing <= cursor && cursor <= len(t.template) {
		t.elements = append(t.elements, Text(t.template[t.lastClosing:cursor]))
	}
}

func (t *_tokenizer) tryOpenMustache() bool {
	if t.prevChar == '\\' {
		return false
	}

	if t.char == '{' && (t.nextChar == '{' || t.nextChar == '#' || t.nextChar == '!') {
		t.grabText(t.cursor)

		t.currentTag = &Mustache{
			start:     t.cursor,
			Location:  helpers.Location{t.line, t.column},
			IsComment: t.nextChar == '#',
			Sanitize:  t.nextChar == '{',
		}
		t.isInsideTag = true

		t.next()
		t.next()
		return true
	}

	return false
}

func (t *_tokenizer) tryCloseMustache() (bool, error) {
	if t.nextChar == '}' && (t.char == '}' || t.char == '#' || t.char == '!') {
		if t.currentTag == nil || t.currentTag.IsComment && t.char != '#' || t.currentTag.Sanitize && t.char != '}' {
			return false, errors2.NewError("unexpected tag termination")
		}
		t.currentTag.Literal = string(t.template[t.currentTag.start : t.cursor+2])
		t.lastClosing = t.cursor + 2
		t.elements = append(t.elements, t.currentTag)
		t.currentTag = nil

		t.next()
		t.next()

		t.isInsideTag = false

		return true, nil
	}

	return false, nil
}

func (t *_tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.char) && t.char != 0 {
		t.next()
	}
}

func (t *_tokenizer) isValidVariableName() bool {
	return unicode.IsLetter(t.char) || t.char == '_'
}

func (t *_tokenizer) isDigit() bool {
	return t.char >= '0' && t.char <= '9'
}

func isAsciiLetter(chr rune) bool {
	return chr >= 'a' && chr <= 'z' || chr >= 'A' && chr <= 'Z' || chr == '_'
}

func isValidStatementLiteral(literal string) bool {
	return slices.Index(Instructions, literal) != -1
}

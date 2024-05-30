package tokenizer

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"math"
	"slices"
	"strings"
	"unicode"
)

type _tokenizer struct {
	filename    string
	rawTemplate string
	template    []rune
	cursor      int

	elements []Element

	lastClosing int

	line   int
	column int
}

type Token struct {
	Kind     string
	Literal  string
	Start    int
	Length   int
	Location helpers.Location
}

func (t Token) String() string {
	return fmt.Sprintf("(%s [%s])", t.Literal, t.Kind)
}

func Tokenize(filename string, template string) ([]Element, error) {
	t := &_tokenizer{
		filename:    filename,
		rawTemplate: template,
		template:    []rune(template),
		elements:    make([]Element, 0),
		cursor:      -1,
		line:        1,
		column:      0,
	}

	return t.tokenize()
}

func (t *_tokenizer) tokenize() ([]Element, error) {
	t.forward()

	for t.rune() != 0 {
		// escaping character
		if t.rune() == '\\' {
			t.forward()
			t.forward()

			continue
		}

		// comment block, ignore it
		if t.rune() == '{' && t.nextRune() == '#' {
			t.grabText(t.cursor)

			t.forward()
			t.forward()

			for t.rune() != '#' && t.nextRune() != '}' {
				if t.nextRune() == 0 {
					if t.rune() != 0 {
						t.forward()
					}
					return nil, t.error("unexpected EOF, unclosed comment", t.location())
				}

				t.forward()
			}

			t.forward()
			t.forward()

			t.lastClosing = t.cursor
		} else if t.rune() == '{' && (t.nextRune() == '{' || t.nextRune() == '!') {
			location := t.location()
			t.grabText(t.cursor)

			sanitize := t.nextRune() != '!'

			t.forward()
			t.forward()

			tokens, err := t.tokenizeExpression(true)
			if err != nil {
				return nil, err
			}

			closing := t.location()
			closingSymbol := fmt.Sprintf("%c%c", t.rune(), t.nextRune())
			t.forward()
			t.forward()

			if sanitize && closingSymbol != "}}" {
				return nil, t.error("expected `}}` to close mustache", closing)
			} else if !sanitize && closingSymbol != "!}" {
				return nil, t.error("expected `!}` to close mustache", closing)
			}

			if len(tokens) == 0 {
				return nil, t.error("empty statement", location)
			}

			t.elements = append(t.elements, &Mustache{
				Sanitize: sanitize,
				Tokens:   tokens,
				Location: location.SetLength(t.cursor - location.Cursor),
			})

			t.lastClosing = t.cursor
		} else if t.rune() == '@' && isAsciiLetter(t.nextRune()) {
			location := t.location()
			t.grabText(t.cursor)

			t.forward()
			start := t.cursor
			for isAsciiLetter(t.rune()) {
				t.forward()
			}
			instruction := string(t.template[start:t.cursor])

			if !slices.Contains(Instructions, instruction) {
				return nil, t.error(fmt.Sprintf("unexpected instruction: `%s`", instruction), location)
			}

			var tokens []Token

			// for those statements that require some arguments
			if !strings.HasPrefix(instruction, "end") && instruction != "else" {
				for t.rune() != '(' && t.rune() != 0 {
					t.forward()
				}
				if t.rune() == 0 {
					return nil, t.error("unexpected EOF, expected `(` after statement", location)
				}

				if t.rune() == '(' {
					t.forward()
					var err error
					tokens, err = t.tokenizeExpression(false)
					if err != nil {
						return nil, err
					}
					t.forward()

					if len(tokens) == 0 {
						return nil, t.error("empty statement", location)
					}
				}
			}

			t.elements = append(t.elements, &Statement{
				Instruction: instruction,
				Tokens:      tokens,
				Location:    location.SetLength(t.cursor - location.Cursor),
			})

			t.lastClosing = t.cursor
		} else {
			t.forward()
		}
	}

	t.grabText(len(t.template))

	return t.elements, nil
}

func (t *_tokenizer) tokenizeExpression(mustache bool) ([]Token, error) {
	parens := helpers.Stack[rune]{}
	tokens := make([]Token, 0)

	for t.rune() != 0 {
		t.skipWhitespace()

		pushNext := true
		token := Token{Start: t.cursor, Length: 1, Literal: string(t.rune()), Location: t.location()}

		switch t.rune() {
		case '?':
			if t.nextRune() == '.' {
				token.Kind = TokOptionalChain
				token.Literal = "?."
				t.forward()
			} else if t.nextRune() == ':' {
				token.Kind = TokElvis
				token.Literal = "?:"
				t.forward()
			} else {
				token.Kind = TokQuestion
			}
		case ':':
			token.Kind = TokColon
		case '%':
			token.Kind = TokModulo
		case '[':
			token.Kind = TokLbrack
			parens.Push('[')
		case ']':
			token.Kind = TokRbrack
			if parens.IsEmpty() {
				return nil, t.error("unexpected `]`", t.location())
			} else if parens.Pop() != '[' {
				return nil, t.error("unexpected `]`, as it closes `(`", t.location())
			}
		case '}':
			if t.nextRune() == '}' {
				return tokens, nil
			}
		case '@':
			token.Kind = TokAt
		case '+':
			token.Kind = TokPlus
		case '-':
			token.Kind = TokMinus
		case '*':
			if t.nextRune() == '*' {
				token.Kind = TokPower
				token.Literal = "**"
				t.forward()
			} else {
				token.Kind = TokAsterisk
			}
		case '/':
			token.Kind = TokSlash
		case '<':
			if t.nextRune() == '=' {
				token.Kind = TokLte
				token.Literal = "<="
				t.forward()
			} else {
				token.Kind = TokLt
			}
		case '>':
			if t.nextRune() == '=' {
				token.Kind = TokGte
				token.Literal = ">="
				t.forward()
			} else {
				token.Kind = TokGt
			}
		case '!':
			if t.nextRune() == '=' {
				token.Kind = TokNeq
				token.Literal = "!="
				t.forward()
			} else if mustache && t.nextRune() == '}' {
				return tokens, nil
			} else {
				token.Kind = TokBang
			}
		case '"', '\'':
			quoteChar := t.rune()
			t.forward()
			previous := t.rune()
			start := t.cursor
			for t.rune() != quoteChar || (t.rune() == quoteChar && previous == '\\') {
				if t.rune() == 0 {
					return nil, t.error("unexpected EOF, unclosed string", t.location())
				}
				previous = t.rune()
				t.forward()
			}
			token.Kind = TokString
			token.Literal = string(t.template[start:t.cursor])
		case ',':
			token.Kind = TokComma
		case '(':
			token.Kind = TokLparen
			parens.Push('(')
		case ')':
			token.Kind = TokRparen
			if parens.IsEmpty() {
				if mustache {
					return nil, t.error("unexpected `)`", t.location())
				} else {
					return tokens, nil
				}
			} else if parens.Pop() != '(' {
				return nil, t.error("unexpected `)`, as it closes `[`", t.location())
			}
		case '&':
			if t.nextRune() == '&' {
				token.Kind = TokAnd
				token.Literal = "&&"
				t.forward()
			}
		case '|':
			if t.nextRune() == '|' {
				token.Kind = TokOr
				token.Literal = "||"
				t.forward()
			}
		case '=':
			if t.nextRune() == '=' {
				token.Kind = TokEq
				token.Literal = "=="
				t.forward()
				break
			}
			fallthrough
		case '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if t.rune() == '.' && !isDigit(t.nextRune(), 10) {
				token.Kind = TokDot
			} else {
				var err error
				token, err = t.numeric()
				if err != nil {
					return nil, err
				}
				pushNext = false
			}
		default:
			if t.isValidVariableStart() {
				start := t.cursor
				for t.isValidVariableStart() || isDigit(t.rune(), 10) {
					t.forward()
				}
				literal := string(t.template[start:t.cursor])
				if slices.Index(Keywords, literal) != -1 {
					token.Kind, token.Literal = literal, literal
				} else {
					token.Kind = TokIdent
					token.Literal = literal
				}
				pushNext = false
			} else if t.rune() == 0 {
				return nil, t.error("unexpected EOF", t.location())
			} else {
				return nil, t.error(fmt.Sprintf("unexpected token: `%c`", t.rune()), t.location())
			}
		}

		token.Length = t.cursor - token.Start
		if pushNext {
			t.forward()
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (t *_tokenizer) numeric() (token Token, err error) {
	token.Location = t.location()

	radix := 10
	start := t.cursor
	if t.rune() == '0' {
		t.forward()

		switch t.rune() {
		case 'x', 'X':
			t.forward()
			radix = 16
		case 'b', 'B':
			t.forward()
			radix = 2
		case 'o', 'O':
			t.forward()
			radix = 8
		default:
			if isDigit(t.rune(), 10) {
				radix = 8
			}
		}
	}

	for isDigit(t.rune(), radix) || (isDigit(t.previousRune(), radix) && t.rune() == '_' && isDigit(t.nextRune(), radix)) {
		t.forward()
	}

	if t.rune() == '.' && radix != 10 {
		return token, t.error("unexpected floating point number in non decimal literal", t.location())
	}

	if t.rune() == '.' && radix == 10 {
		t.forward()
		for isDigit(t.rune(), radix) || (isDigit(t.previousRune(), radix) && t.rune() == '_' && isDigit(t.nextRune(), radix)) {
			t.forward()
		}
	}

	token.Kind = TokNumeric
	token.Literal = strings.ReplaceAll(string(t.template[start:t.cursor]), "_", "")
	token.Start = start
	token.Length = t.cursor - start

	if isLetter(t.rune()) || isDigit(t.rune(), 10) || t.rune() == '_' || t.rune() == '.' {
		return token, t.error(fmt.Sprintf("unexpected character `%s` in numeric literal", string(t.rune())), t.location())
	}

	return
}

func (t *_tokenizer) grabText(cursor int) {
	bounded := int(math.Min(float64(cursor), float64(len(t.template))))
	if t.lastClosing < bounded && t.lastClosing-bounded != 0 {
		t.elements = append(t.elements, Text(t.template[t.lastClosing:bounded]))
	}
}

func (t *_tokenizer) location() helpers.Location {
	return helpers.Location{t.line, t.column, t.cursor, 1}
}

func (t *_tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.rune()) {
		t.forward()
	}
}

func (t *_tokenizer) rune() rune {
	if t.cursor >= len(t.template) {
		return 0
	}
	return t.template[t.cursor]
}

func (t *_tokenizer) nextRune() rune {
	if t.cursor+1 >= len(t.template) {
		return 0
	}
	return t.template[t.cursor+1]
}

func (t *_tokenizer) previousRune() rune {
	if t.cursor-1 < 0 {
		return 0
	}
	return t.template[t.cursor-1]
}

func (t *_tokenizer) forward() {
	t.cursor++
	if t.rune() == '\n' {
		t.line++
		t.column = 0
	} else {
		t.column++
	}
}

func isDigit(r rune, radix int) bool {
	if radix == 10 {
		return r >= '0' && r <= '9'
	}

	if radix == 16 {
		return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
	}

	if radix == 8 {
		return r >= '0' && r <= '7'
	}

	if radix == 2 {
		return r == '0' || r == '1'
	}

	return false
}

func isLetter(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isAsciiLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func (t *_tokenizer) isValidVariableStart() bool {
	return isLetter(t.rune())
}

func (t *_tokenizer) error(message string, start helpers.Location) error {
	return errors2.New(message, t.filename, t.rawTemplate, start, t.location())
}

package tokenizer

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

type _tokenizer struct {
	templateRunes []rune
	template      string
	elements      []Element

	lastClosing int

	cursor    int
	rawCursor int

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

func Tokenize(template string) ([]Element, error) {
	t := &_tokenizer{
		templateRunes: []rune(template),
		template:      template,
		elements:      make([]Element, 0),
		cursor:        0,
		rawCursor:     0,
		line:          1,
		column:        0,
	}

	return t.tokenize()
}

func (t *_tokenizer) tokenize() ([]Element, error) {
	possibleElements := lookupElements(t.template)
	if len(possibleElements) == 0 {
		t.elements = append(t.elements, Text(t.template))
		return t.elements, nil
	}

	for _, element := range possibleElements {
		if element[0] < t.rawCursor {
			continue
		}

		t.goTo(element[0])
		t.grabText(t.cursor)

		if t.rune() == '{' && t.nextRune() == '#' {
			t.forward()
			t.forward()

			location := t.location()

			for t.rune() != '#' && t.nextRune() != '}' {
				if t.nextRune() == 0 {
					return nil, errors2.New("unexpected end of template, unclosed comment", location)
				}

				t.forward()
			}

			t.forward()
			t.forward()
		} else if t.rune() == '{' && (t.nextRune() == '{' || t.nextRune() == '!') {
			t.forward()

			sanitize := t.rune() != '!'

			t.forward()

			location := helpers.Location{
				Line:   t.line,
				Column: t.column,
			}

			tokens, err := t.tokenizeExpression(true, sanitize)
			if err != nil {
				return nil, err
			}

			t.elements = append(t.elements, &Mustache{
				Literal:  t.template[element[0]+2 : t.rawCursor],
				Sanitize: sanitize,
				Tokens:   tokens,
				Location: location,
			})

			t.forward()
			t.forward()
		} else {
			var tokens []Token
			var err error

			location := t.location()
			t.forward()
			instruction := t.template[element[0]+1 : element[1]]
			t.goTo(element[1])

			if t.rune() != '(' && !strings.HasPrefix(instruction, "end") {
				return nil, errors2.New("expected `(` after statement", location)
			}

			if t.rune() == '(' {
				t.forward()
				tokens, err = t.tokenizeExpression(false, false)
				if err != nil {
					return nil, err
				}
				t.forward()
			}

			t.elements = append(t.elements, &Statement{
				Literal:     t.template[element[0]:t.rawCursor],
				Instruction: instruction,
				Tokens:      tokens,
				Location:    location,
			})

		}

		t.lastClosing = t.cursor
	}

	t.grabText(len(t.templateRunes))

	return t.elements, nil
}

func (t *_tokenizer) tokenizeExpression(mustache bool, sanitizedMustache bool) ([]Token, error) {
	parens := helpers.Stack[rune]{}
	tokens := make([]Token, 0)

	for t.rune() != 0 {
		t.skipWhitespace()

		pushNext := true
		token := Token{Start: t.cursor, Length: 1, Literal: string(t.rune()), Location: t.location()}

		switch t.rune() {
		case '.':
			token.Kind = TokDot
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
			if parens.Pop() != '[' {
				return nil, errors2.New("unexpected closing bracket", t.location())
			}
		case '}':
			if sanitizedMustache && t.nextRune() == '}' {
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
			} else if mustache && !sanitizedMustache && t.nextRune() == '}' {
				return tokens, nil
			} else {
				token.Kind = TokBang
			}
		case '"', '\'':
			quoteChar := t.rune()
			t.forward()
			previous := t.rune()
			start := t.rawCursor
			for t.rune() != quoteChar || (t.rune() == quoteChar && previous == '\\') {
				if t.rune() == 0 {
					return nil, errors2.New("unexpected end of template, unclosed string", t.location())
				}
				previous = t.rune()
				t.forward()
			}
			literal := t.template[start:t.rawCursor]
			t.forward()
			token.Kind = TokString
			token.Literal = literal
			pushNext = false
		case ',':
			token.Kind = TokComma
		case '(':
			token.Kind = TokLparen
			parens.Push('(')
		case ')':
			token.Kind = TokRparen
			if parens.Pop() != '(' {
				if mustache {
					return nil, errors2.New("unexpected closing parenthesis", t.location())
				} else {
					return tokens, nil
				}
			}
		case '=':
			if t.nextRune() == '=' {
				token.Kind = TokEq
				token.Literal = "=="
				t.forward()
				break
			}
			fallthrough
		default:
			if t.isValidVariableStart() {
				start := t.rawCursor
				for t.isValidVariableStart() || t.isDigit() {
					t.forward()
				}
				literal := t.template[start:t.rawCursor]
				if slices.Index(Keywords, literal) != -1 {
					token.Kind, token.Literal = literal, literal
				} else {
					token.Kind = TokIdent
					token.Literal = literal
				}
				pushNext = false
			} else if t.isDigit() {
				start := t.rawCursor
				hasDot := false
				for t.isDigit() || t.rune() == '.' {
					t.forward()
					if t.rune() == '.' {
						if hasDot {
							return nil, errors2.New("unexpected dot in number", t.location())
						}
						hasDot = true
					}
				}
				token.Kind = TokNumber
				token.Literal = t.template[start:t.rawCursor]
				pushNext = false
			} else {
				return nil, errors2.New(fmt.Sprintf("unexpected token: '%s'", string(t.rune())), t.location())
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

func (t *_tokenizer) grabText(cursor int) {
	bounded := int(math.Min(float64(cursor), float64(len(t.templateRunes))))
	if t.lastClosing < bounded && t.lastClosing-bounded != 0 {
		t.elements = append(t.elements, Text(t.templateRunes[t.lastClosing:bounded]))
	}
}

func (t *_tokenizer) location() helpers.Location {
	return helpers.Location{
		Line:   t.line,
		Column: t.column,
	}
}

func (t *_tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.rune()) {
		t.forward()
	}
}

func (t *_tokenizer) rune() rune {
	if t.cursor >= len(t.templateRunes) {
		return 0
	}
	return t.templateRunes[t.cursor]
}

func (t *_tokenizer) nextRune() rune {
	if t.cursor+1 >= len(t.templateRunes) {
		return 0
	}
	return t.templateRunes[t.cursor+1]
}

func (t *_tokenizer) goTo(location int) {
	rawCursor := t.rawCursor
	for i := range t.template[rawCursor:] {
		if t.rune() == '\n' {
			t.line++
			t.column = 0
		} else {
			t.column++
		}

		if i+rawCursor == location {
			t.rawCursor = i + rawCursor
			break
		}
		t.cursor++

	}
}

func (t *_tokenizer) forward() {
	t.rawCursor += len(string(t.rune()))
	t.cursor++
	if t.rune() == '\n' {
		t.line++
		t.column = 0
	} else {
		t.column++
	}
}

func (t *_tokenizer) isDigit() bool {
	return t.rune() >= '0' && t.rune() <= '9'
}

func (t *_tokenizer) isValidVariableStart() bool {
	return t.rune() == '_' || unicode.IsLetter(t.rune())
}

var lookupRegex *regexp.Regexp

func generateLookupRegex() {
	pattern := fmt.Sprintf(`(?:^|[^\\])(\{[\{#!]|@(?:%s))`, strings.Join(Instructions, "|"))
	lookupRegex = regexp.MustCompile(pattern)
}

func lookupElements(haystack string) [][]int {
	if lookupRegex == nil {
		generateLookupRegex()
	}

	found := lookupRegex.FindAllStringIndex(haystack, -1)

	for i, element := range found {
		if element[1]-element[0] == 2 {
			continue
		}

		if haystack[element[0]] == haystack[element[0]+1] {
			found = append(found[:i], found[i+1:]...)
			continue
		}

		actualOpeningAt := strings.IndexAny(haystack[element[0]:element[1]], "{@")
		if actualOpeningAt != 0 || haystack[element[0]+1] == '@' || haystack[element[0]+1] == '{' {
			if haystack[element[0]] == '@' || haystack[element[0]] == '{' {
				element[0]++
				continue
			}
			element[0] += actualOpeningAt
		}
	}

	return found
}

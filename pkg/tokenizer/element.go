package tokenizer

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"strings"
)

type ElementKind string

const (
	TextKind      ElementKind = "text"
	MustacheKind  ElementKind = "mustache"
	StatementKind ElementKind = "statement"
)

type Element interface {
	Kind() ElementKind
	String() string
}

type Text string

func (t Text) Kind() ElementKind {
	return TextKind
}

func (t Text) String() string {
	return fmt.Sprintf("TEXT: `%s`", strings.ReplaceAll(string(t), "\n", "\\n"))
}

func (t Text) Tokens() []Token {
	return nil
}

type Mustache struct {
	start     int
	Literal   string
	Sanitize  bool
	IsComment bool
	Tokens    []Token
	Location  helpers.Location
}

func (t *Mustache) Kind() ElementKind {
	return MustacheKind
}

func (t *Mustache) String() string {
	return t.Literal
}

type Statement struct {
	Instruction string
	Tokens      []Token
	Flags       []string
	Location    helpers.Location
}

func (s *Statement) Kind() ElementKind {
	return StatementKind
}

func (s *Statement) String() string {
	return s.Instruction
}

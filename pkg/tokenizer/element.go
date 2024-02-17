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
	return fmt.Sprintf("TEXT      : `%s`", strings.ReplaceAll(string(t), "\n", "\\n"))
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
	return fmt.Sprintf("MUSTACHE  : `%s`", t.Literal)
}

type Statement struct {
	Literal     string
	Instruction string
	Tokens      []Token
	Flags       []string
	Location    helpers.Location
}

func (s *Statement) Kind() ElementKind {
	return StatementKind
}

func (s *Statement) String() string {
	if s.Literal == "" {
		return fmt.Sprintf("STATEMENT : `@%s`", s.Instruction)
	}
	return fmt.Sprintf("STATEMENT : `@%s%s`", s.Instruction, s.Literal)
}

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
}

type Text string

func (t Text) Kind() ElementKind {
	return TextKind
}

func (t Text) String() string {
	if len(t) > 80 {
		return fmt.Sprintf(
			"TEXT(%s [...] %s)",
			strings.ReplaceAll(string(t[:40]), "\n", "\\n"),
			strings.ReplaceAll(string(t[len(t)-40:]), "\n", "\\n"),
		)
	}

	return fmt.Sprintf("TEXT(%s)", strings.ReplaceAll(string(t), "\n", "\\n"))
}

type Mustache struct {
	Literal  string
	Sanitize bool
	Tokens   []Token
	Location helpers.Location
}

func (t *Mustache) Kind() ElementKind {
	return MustacheKind
}

func (t *Mustache) String() string {
	return fmt.Sprintf("MUSTACHE(%s)", t.Literal)
}

type Statement struct {
	Literal     string
	Instruction string
	Tokens      []Token
	Location    helpers.Location
}

func (s *Statement) Kind() ElementKind {
	return StatementKind
}

func (s *Statement) String() string {
	return fmt.Sprintf("STATEMENT(%s)", s.Literal)
}

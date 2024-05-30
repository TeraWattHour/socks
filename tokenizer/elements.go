package tokenizer

import (
	"github.com/terawatthour/socks/internal/helpers"
)

type ElementKind int

const (
	TextKind ElementKind = iota
	MustacheKind
	StatementKind
)

type Element interface {
	Kind() ElementKind
}

type Text string

func (t Text) Kind() ElementKind {
	return TextKind
}

type Mustache struct {
	Sanitize bool
	Tokens   []Token
	Location helpers.Location
}

func (t *Mustache) Kind() ElementKind {
	return MustacheKind
}

type Statement struct {
	Instruction string
	Tokens      []Token
	Location    helpers.Location
}

func (s *Statement) Kind() ElementKind {
	return StatementKind
}

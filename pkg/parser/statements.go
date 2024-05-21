package parser

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/expression"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"strings"
)

type Program interface {
	Kind() string
	String() string
	Location() helpers.Location
}

type Text struct {
	Content string
}

func (t *Text) Kind() string {
	return "text"
}

func (t *Text) String() string {
	return fmt.Sprintf("%-8s: `%s`", "TEXT", strings.ReplaceAll(t.Content, "\n", "\\n"))
}

func (t *Text) Tag() *tokenizer.Mustache {
	return nil
}

func (t *Text) Location() helpers.Location {
	panic("unreachable")
}

// ---------------------- Expression (Mustache) ----------------------

type Expression struct {
	Program      *expression.VM
	tag          *tokenizer.Mustache
	Dependencies []string
}

func (vs *Expression) Location() helpers.Location {
	return vs.tag.Location
}

func (vs *Expression) String() string {
	return fmt.Sprintf("%-8s: %s", "MUSTACHE", vs.tag.Literal)
}

func (vs *Expression) Kind() string {
	return "expression"
}

func (vs *Expression) Tag() *tokenizer.Mustache {
	return vs.tag
}

type Statement = Program

// ---------------------- If Statement ----------------------

type IfStatement struct {
	Program      *expression.VM
	location     helpers.Location
	Dependencies []string
	EndStatement Statement
}

func (vs *IfStatement) Location() helpers.Location {
	return vs.location
}

func (vs *IfStatement) String() string {
	return fmt.Sprintf("%-8s", "IF")
}

func (vs *IfStatement) Kind() string {
	return "if"
}

// ---------------------- For Statement ----------------------

type ForStatement struct {
	Iterable     *expression.VM
	KeyName      string
	ValueName    string
	location     helpers.Location
	Dependencies []string
	EndStatement *EndStatement
}

func (es *ForStatement) Location() helpers.Location {
	return es.location
}

func (es *ForStatement) String() string {
	if es.KeyName != "" {
		return fmt.Sprintf("%-8s: %s, %s in [%p]", "FOR", es.KeyName, es.ValueName, es)
	}
	return fmt.Sprintf("%-8s: %s in [%p]", "FOR", es.ValueName, es)
}

func (es *ForStatement) Kind() string {
	return "for"
}

// ---------------------- Extend Statement ----------------------

type ExtendStatement struct {
	Template string
	location helpers.Location
}

func (es *ExtendStatement) Location() helpers.Location {
	return es.location
}

func (es *ExtendStatement) String() string {
	return fmt.Sprintf("%-8s: %s", "EXTEND", es.Template)
}

func (es *ExtendStatement) Kind() string {
	return "extend"
}

// ---------------------- template Statement ----------------------

type TemplateStatement struct {
	Template     string
	location     helpers.Location
	EndStatement *EndStatement
}

func (es *TemplateStatement) Location() helpers.Location {
	return es.location
}

func (es *TemplateStatement) String() string {
	return fmt.Sprintf("%-8s: %s", "TEMPLATE", es.Template)
}

func (es *TemplateStatement) Kind() string {
	return "template"
}

// ---------------------- Slot Statement ----------------------

type SlotStatement struct {
	Name         string
	location     helpers.Location
	Parent       Statement
	EndStatement *EndStatement
}

func (ss *SlotStatement) Location() helpers.Location {
	return ss.location
}

func (ss *SlotStatement) String() string {
	return fmt.Sprintf("%-8s: %s", "SLOT", ss.Name)
}

func (ss *SlotStatement) Kind() string {
	return "slot"
}

// ---------------------- Define Statement ----------------------

type DefineStatement struct {
	Name         string
	location     helpers.Location
	Parent       Statement
	EndStatement *EndStatement
}

func (es *DefineStatement) Location() helpers.Location {
	return es.location
}

func (es *DefineStatement) Kind() string {
	return "define"
}

func (es *DefineStatement) String() string {
	return fmt.Sprintf("%-8s: %s", "DEFINE", es.Name)
}

type EndStatement struct {
	location        helpers.Location
	ClosedStatement Statement
}

func (es *EndStatement) Location() helpers.Location {
	return es.ClosedStatement.Location()
}

func (es *EndStatement) Kind() string {
	return "end"
}

func (es *EndStatement) String() string {
	return fmt.Sprintf("END(%s)", es.ClosedStatement.Kind())
}

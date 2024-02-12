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
	ChangeProgramCount(int)
	SetParent(Statement)
}

type Text struct {
	Content string
	Parent  Statement
}

func (t *Text) ChangeProgramCount(i int) {
	if t.Parent != nil {
		t.Parent.ChangeProgramCount(i)
	}
}

func (t *Text) Kind() string {
	return "text"
}

func (t *Text) String() string {
	return fmt.Sprintf("%s: `%s`", helpers.FixedWidth("TEXT", 8), strings.ReplaceAll(t.Content, "\n", "\\n"))
}

func (t *Text) Tag() *tokenizer.Mustache {
	return nil
}

func (t *Text) Location() helpers.Location {
	panic("unreachable")
}

func (t *Text) SetParent(p Statement) {
	t.Parent = p
}

type Statement interface {
	Kind() string
	NoStatic() bool
	String() string
	ChangeProgramCount(int)
	Location() helpers.Location
	SetParent(Statement)
}

// ---------------------- Print Statement ----------------------

type PrintStatement struct {
	Program      *expression.VM
	tag          *tokenizer.Mustache
	noStatic     bool
	Parent       Statement
	Dependencies []string
}

func (vs *PrintStatement) SetParent(p Statement) {
	vs.Parent = p
}

func (vs *PrintStatement) Location() helpers.Location {
	return vs.tag.Location
}

func (vs *PrintStatement) ChangeProgramCount(i int) {
	if vs.Parent != nil {
		vs.Parent.ChangeProgramCount(i)
	}
}

func (vs *PrintStatement) String() string {
	return "print"
}

func (vs *PrintStatement) NoStatic() bool {
	return vs.noStatic
}

func (vs *PrintStatement) Kind() string {
	return "variable"
}

func (vs *PrintStatement) Tag() *tokenizer.Mustache {
	return vs.tag
}

// ---------------------- If Statement ----------------------

type IfStatement struct {
	Program      *expression.VM
	bodyStart    int
	Programs     int
	noStatic     bool
	Parent       Statement
	location     helpers.Location
	Dependencies []string
}

func (vs *IfStatement) SetParent(p Statement) {
	vs.Parent = p
}

func (vs *IfStatement) Location() helpers.Location {
	return vs.location
}

func (vs *IfStatement) ChangeProgramCount(i int) {
	vs.Programs += i
	if vs.Parent != nil {
		vs.Parent.ChangeProgramCount(i)
	}
}

func (vs *IfStatement) NoStatic() bool {
	return vs.noStatic
}

func (vs *IfStatement) String() string {
	return "if"
}

func (vs *IfStatement) Kind() string {
	return "if"
}

// ---------------------- For Statement ----------------------

type ForStatement struct {
	Iterable     *expression.VM
	KeyName      string
	ValueName    string
	Programs     int
	bodyStart    int
	noStatic     bool
	Parent       Statement
	location     helpers.Location
	Dependencies []string
}

func (es *ForStatement) SetParent(p Statement) {
	es.Parent = p
}

func (es *ForStatement) Location() helpers.Location {
	return es.location
}

func (es *ForStatement) ChangeProgramCount(i int) {
	es.Programs += i
	if es.Parent != nil {
		es.Parent.ChangeProgramCount(i)
	}
}

func (es *ForStatement) String() string {
	if es.KeyName != "" {
		return fmt.Sprintf("%s: %s, %s in [%p]", helpers.FixedWidth("FOR", 8), es.KeyName, es.ValueName, es)
	}
	return fmt.Sprintf("%s: %s in [%p]", helpers.FixedWidth("FOR", 8), es.ValueName, es)
}

func (es *ForStatement) NoStatic() bool {
	return es.noStatic
}

func (es *ForStatement) Kind() string {
	return "for"
}

// ---------------------- Extend Statement ----------------------

type ExtendStatement struct {
	Template string
	location helpers.Location
}

func (es *ExtendStatement) SetParent(p Statement) {
	return
}

func (es *ExtendStatement) Location() helpers.Location {
	return es.location
}

func (es *ExtendStatement) ChangeProgramCount(i int) {
	return
}

func (es *ExtendStatement) String() string {
	return fmt.Sprintf("extend: %s", es.Template)
}

func (es *ExtendStatement) NoStatic() bool {
	return false
}

func (es *ExtendStatement) Kind() string {
	return "extend"
}

// ---------------------- template Statement ----------------------

type TemplateStatement struct {
	Template  string
	location  helpers.Location
	Programs  int
	BodyStart int
	Depth     int
	Parent    Statement
}

func (es *TemplateStatement) SetParent(p Statement) {
	es.Parent = p
}

func (es *TemplateStatement) Location() helpers.Location {
	return es.location
}

func (es *TemplateStatement) ChangeProgramCount(i int) {
	es.Programs += i
	if es.Parent != nil {
		es.Parent.ChangeProgramCount(i)
	}
}

func (es *TemplateStatement) String() string {
	return fmt.Sprintf("%s: %s", helpers.FixedWidth("TEMPLATE", 8), es.Template)
}

func (es *TemplateStatement) NoStatic() bool {
	return false
}

func (es *TemplateStatement) Kind() string {
	return "template"
}

// ---------------------- Slot Statement ----------------------

type SlotStatement struct {
	Name      string
	Programs  int
	bodyStart int
	Depth     int
	Parent    Statement
	location  helpers.Location
}

func (ss *SlotStatement) SetParent(p Statement) {
	ss.Parent = p
}

func (ss *SlotStatement) Location() helpers.Location {
	return ss.location
}

func (ss *SlotStatement) ChangeProgramCount(i int) {
	ss.Programs += i
	if ss.Parent != nil {
		ss.Parent.ChangeProgramCount(i)
	}
}

func (ss *SlotStatement) String() string {
	return fmt.Sprintf("slot: %s", ss.Name)
}

func (ss *SlotStatement) NoStatic() bool {
	return false
}

func (ss *SlotStatement) Kind() string {
	return "slot"
}

// ---------------------- Define Statement ----------------------

type DefineStatement struct {
	Name      string
	location  helpers.Location
	Programs  int
	bodyStart int
	Parent    Statement
	Depth     int
}

func (es *DefineStatement) SetParent(p Statement) {
	es.Parent = p
}

func (es *DefineStatement) Location() helpers.Location {
	return es.location
}

func (es *DefineStatement) ChangeProgramCount(i int) {
	es.Programs += i
	if es.Parent != nil {
		es.Parent.ChangeProgramCount(i)
	}
}

func (es *DefineStatement) Kind() string {
	return "define"
}

func (es *DefineStatement) String() string {
	return fmt.Sprintf("%s: %s", helpers.FixedWidth("DEFINE", 8), es.Name)
}

func (es *DefineStatement) NoStatic() bool {
	return false
}

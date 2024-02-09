package parser

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/expression"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"strings"
)

var PreprocessorKinds = []string{"extend", "template", "slot", "define"}

type Program interface {
	Kind() string
	String() string
	Tag() *tokenizer.Tag
}

type Text string

func (t Text) Kind() string {
	return "text"
}

func (t Text) String() string {
	return fmt.Sprintf("%s: `%s`", helpers.FixedWidth("TEXT", 8), strings.ReplaceAll(string(t), "\n", "\\n"))
}

func (t Text) Tag() *tokenizer.Tag {
	return nil
}

type Statement interface {
	Kind() string
	Tag() *tokenizer.Tag
	NoStatic() bool
	String() string
	ChangeProgramCount(int)
}

// ---------------------- Print Statement ----------------------

type PrintStatement struct {
	Program  *expression.VM
	tag      *tokenizer.Tag
	noStatic bool
}

func (vs *PrintStatement) ChangeProgramCount(i int) {
	return
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

func (vs *PrintStatement) Tag() *tokenizer.Tag {
	return vs.tag
}

// ---------------------- If Statement ----------------------

type IfStatement struct {
	Program   *expression.VM
	StartTag  *tokenizer.Tag
	EndTag    *tokenizer.Tag
	bodyStart int
	Programs  int
	noStatic  bool
	Parent    Statement
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

func (vs *IfStatement) Tag() *tokenizer.Tag {
	return vs.StartTag
}

// ---------------------- For Statement ----------------------

type ForStatement struct {
	Iterable  *expression.VM
	KeyName   string
	ValueName string
	Programs  int
	tag       *tokenizer.Tag
	bodyStart int
	noStatic  bool
	Parent    Statement
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

func (es *ForStatement) Tag() *tokenizer.Tag {
	return es.tag
}

// ---------------------- Extend Statement ----------------------

type ExtendStatement struct {
	Template string
	tag      *tokenizer.Tag
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

func (es *ExtendStatement) Tag() *tokenizer.Tag {
	return es.tag
}

// ---------------------- Template Statement ----------------------

type TemplateStatement struct {
	Template  string
	StartTag  *tokenizer.Tag
	EndTag    *tokenizer.Tag
	Programs  int
	BodyStart int
	Depth     int
	Parent    Statement
}

func (es *TemplateStatement) ChangeProgramCount(i int) {
	es.Programs += i
	if es.Parent != nil {
		es.Parent.ChangeProgramCount(i)
	}
}

func (es *TemplateStatement) String() string {
	return fmt.Sprintf("template: %s", es.Template)
}

func (es *TemplateStatement) NoStatic() bool {
	return false
}

func (es *TemplateStatement) Kind() string {
	return "template"
}

func (es *TemplateStatement) Tag() *tokenizer.Tag {
	return es.StartTag
}

// ---------------------- Slot Statement ----------------------

type SlotStatement struct {
	Name      string
	tag       *tokenizer.Tag
	Programs  int
	bodyStart int
	Depth     int
	Parent    Statement
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

func (ss *SlotStatement) Tag() *tokenizer.Tag {
	return ss.tag
}

func (ss *SlotStatement) Kind() string {
	return "slot"
}

// ---------------------- Define Statement ----------------------

type DefineStatement struct {
	Name      string
	tag       *tokenizer.Tag
	Programs  int
	bodyStart int
	Parent    Statement
	Depth     int
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

func (es *DefineStatement) Tag() *tokenizer.Tag {
	return es.tag
}

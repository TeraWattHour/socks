package parser

import (
	"github.com/antonmedv/expr/vm"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

type Statement interface {
	Kind() string
	Tag() *tokenizer.Tag
	Start() int
	End() int
	Replace(inner []rune, offset int, content []rune) []rune
}

type IfStatement struct {
	Program  *vm.Program
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	parents  []Statement
	Body     []rune
}

func (vs *IfStatement) Replace(inner []rune, offset int, content []rune) []rune {
	leading := string(content[:vs.StartTag.Start+offset])
	trailing := string(content[vs.EndTag.End+1+offset:])
	innerStr := string(inner)
	return []rune(leading + innerStr + trailing)
}

func (vs *IfStatement) Start() int {
	return vs.StartTag.Start
}

func (vs *IfStatement) End() int {
	return vs.EndTag.End
}

func (vs *IfStatement) Kind() string {
	return "if"
}

func (vs *IfStatement) Tag() *tokenizer.Tag {
	return vs.StartTag
}

type VariableStatement struct {
	Program *vm.Program
	tag     *tokenizer.Tag
	parents []Statement
}

func (vs *VariableStatement) Replace(inner []rune, offset int, content []rune) []rune {
	leading := string(content[:vs.tag.Start+offset])
	trailing := string(content[vs.tag.End+1+offset:])
	innerStr := string(inner)
	return []rune(leading + innerStr + trailing)
}

func (vs *VariableStatement) Start() int {
	return vs.tag.Start
}

func (vs *VariableStatement) End() int {
	return vs.tag.End
}

func (vs *VariableStatement) Kind() string {
	return "variable"
}

func (vs *VariableStatement) Tag() *tokenizer.Tag {
	return vs.tag
}

type ExtendStatement struct {
	Template string
	tag      *tokenizer.Tag
	parents  []Statement
}

func (es *ExtendStatement) Replace(inner []rune, offset int, content []rune) []rune {
	panic("implement me")
}

func (es *ExtendStatement) Start() int {
	return es.tag.Start
}

func (es *ExtendStatement) End() int {
	return es.tag.End
}

func (es *ExtendStatement) Kind() string {
	return "extend"
}

func (es *ExtendStatement) Tag() *tokenizer.Tag {
	return es.tag
}

type TemplateStatement struct {
	Template string
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	parents  []Statement
}

func (es *TemplateStatement) Replace(inner []rune, offset int, content []rune) []rune {
	panic("implement me")
}

func (es *TemplateStatement) Start() int {
	return es.StartTag.Start
}

func (es *TemplateStatement) End() int {
	return es.EndTag.End
}

func (es *TemplateStatement) Kind() string {
	return "template"
}

func (es *TemplateStatement) Tag() *tokenizer.Tag {
	return es.StartTag
}

type SlotStatement struct {
	Name     string
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	parents  []Statement
}

func (es *SlotStatement) Replace(inner []rune, offset int, content []rune) []rune {
	panic("implement me")
}

func (bs *SlotStatement) Start() int {
	return bs.StartTag.Start
}

func (bs *SlotStatement) End() int {
	return bs.EndTag.End
}

func (bs *SlotStatement) Kind() string {
	return "slot"
}

func (bs *SlotStatement) Tag() *tokenizer.Tag {
	return bs.StartTag
}

type EndStatement struct {
	tag     *tokenizer.Tag
	closes  Statement
	parents []Statement
}

func (vs *EndStatement) Replace(inner []rune, offset int, content []rune) []rune {
	leading := string(content[:vs.tag.Start+offset])
	trailing := string(content[vs.tag.End+1+offset:])
	return []rune(leading + trailing)
}

func (es *EndStatement) Start() int {
	return es.tag.Start
}

func (es *EndStatement) End() int {
	return es.tag.End
}

func (es *EndStatement) Kind() string {
	return "end"
}

func (es *EndStatement) Tag() *tokenizer.Tag {
	return es.tag
}

type DefineStatement struct {
	Name     string
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	Parents  []Statement
}

func (es *DefineStatement) Replace(inner []rune, offset int, content []rune) []rune {
	panic("implement me")
}

func (es *DefineStatement) Start() int {
	return es.StartTag.Start
}

func (es *DefineStatement) End() int {
	return es.EndTag.End
}

func (es *DefineStatement) Kind() string {
	return "define"
}

func (es *DefineStatement) Tag() *tokenizer.Tag {
	return es.StartTag
}

type ForStatement struct {
	IteratorName string
	ValueName    string
	Iterable     Statement
	Body         []rune

	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	parents  []Statement
}

func (es *ForStatement) Replace(inner []rune, offset int, content []rune) []rune {
	leading := string(content[:es.StartTag.Start+offset])
	trailing := string(content[es.EndTag.End+1+offset:])
	innerStr := string(inner)
	return []rune(leading + innerStr + trailing)
}

func (es *ForStatement) Start() int {
	return es.StartTag.Start
}

func (es *ForStatement) End() int {
	return es.EndTag.End
}

func (es *ForStatement) Kind() string {
	return "for"
}

func (es *ForStatement) Tag() *tokenizer.Tag {
	return es.StartTag
}

package parser

import "github.com/terawatthour/socks/pkg/tokenizer"

type Statement interface {
	Kind() string
	Tag() *tokenizer.Tag
}

// StringStatement is a string constant embedded in template
type StringStatement struct {
	Value   string
	tag     *tokenizer.Tag
	parents []Statement
}

func (ss *StringStatement) Kind() string {
	return "string"
}

func (ss *StringStatement) Tag() *tokenizer.Tag {
	return ss.tag
}

// IntegerStatement is an integer constant embedded in template
type IntegerStatement struct {
	Value   int
	tag     *tokenizer.Tag
	parents []Statement
}

func (ns *IntegerStatement) Kind() string {
	return "integer"
}

func (ns *IntegerStatement) Tag() *tokenizer.Tag {
	return ns.tag
}

// FloatStatement is a float64 constant embedded in template
type FloatStatement struct {
	Value   float64
	tag     *tokenizer.Tag
	parents []Statement
}

func (fs *FloatStatement) Kind() string {
	return "float"
}

func (fs *FloatStatement) Tag() *tokenizer.Tag {
	return fs.tag
}

// VariableStatement is a chain of context accessors embedded in template,
// also include function calls
type VariableStatement struct {
	Parts   []Statement
	IsLocal bool
	tag     *tokenizer.Tag
	parents []Statement
}

func (vs *VariableStatement) Kind() string {
	return "variable"
}

func (vs *VariableStatement) Tag() *tokenizer.Tag {
	return vs.tag
}

// VariablePartStatement is a part of VariableStatement, it stores
// the name of each part of the variable
type VariablePartStatement struct {
	Name string
	tag  *tokenizer.Tag
}

func (vp *VariablePartStatement) Kind() string {
	return "variable_part"
}

func (vp *VariablePartStatement) Tag() *tokenizer.Tag {
	return vp.tag
}

// FunctionCallStatement stores the arguments of a function call,
// it is a part of VariableStatement
type FunctionCallStatement struct {
	Args []Statement
	tag  *tokenizer.Tag
}

func (vs *FunctionCallStatement) Kind() string {
	return "function_call"
}

func (vs *FunctionCallStatement) Tag() *tokenizer.Tag {
	return vs.tag
}

type ExtendStatement struct {
	Template string
	tag      *tokenizer.Tag
	parents  []Statement
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

	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
	parents  []Statement
}

func (es *ForStatement) Kind() string {
	return "for"
}

func (es *ForStatement) Tag() *tokenizer.Tag {
	return es.StartTag
}

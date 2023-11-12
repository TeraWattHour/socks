package parser

import "github.com/terawatthour/socks/pkg/tokenizer"

type Statement interface {
	Kind() string
}

// StringStatement is a string constant embedded in template
type StringStatement struct {
	Value string
}

func (ss *StringStatement) Kind() string {
	return "string"
}

// IntegerStatement is an integer constant embedded in template
type IntegerStatement struct {
	Value int
}

func (ns *IntegerStatement) Kind() string {
	return "integer"
}

// FloatStatement is a float64 constant embedded in template
type FloatStatement struct {
	Value float64
}

func (fs *FloatStatement) Kind() string {
	return "float"
}

// VariableStatement is a chain of context accessors embedded in template,
// also include function calls
type VariableStatement struct {
	Parts   []Statement
	IsLocal bool
}

func (vs *VariableStatement) Kind() string {
	return "variable"
}

// VariablePartStatement is a part of VariableStatement, it stores
// the name of each part of the variable
type VariablePartStatement struct {
	Name string
}

func (vp *VariablePartStatement) Kind() string {
	return "variable_part"
}

// FunctionCallStatement stores the arguments of a function call,
// it is a part of VariableStatement
type FunctionCallStatement struct {
	Args []Statement
}

func (vs *FunctionCallStatement) Kind() string {
	return "function_call"
}

type ExtendStatement struct {
	Template string
}

func (es *ExtendStatement) Kind() string {
	return "extend"
}

type SlotStatement struct {
	Name     string
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
}

func (bs *SlotStatement) Kind() string {
	return "slot"
}

type EndStatement struct{}

func (es *EndStatement) Kind() string {
	return "end"
}

type DefineStatement struct {
	Name     string
	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
}

func (es *DefineStatement) Kind() string {
	return "define"
}

type ForStatement struct {
	IteratorName string
	ValueName    string
	Iterable     Statement

	StartTag *tokenizer.Tag
	EndTag   *tokenizer.Tag
}

func (es *ForStatement) Kind() string {
	return "for"
}

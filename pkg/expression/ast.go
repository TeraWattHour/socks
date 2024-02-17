package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"strings"
)

type WrappedExpression struct {
	Expr         Expression
	Dependencies []string
}

type Node interface {
	Literal() string
	IsEqual(Node) bool
	String() string
}

type Expression interface {
	Node
	Type() string
}

type Identifier struct {
	Value string
	Token *tokenizer.Token
}

func (s *Identifier) IsEqual(node Node) bool {
	if node, ok := node.(*Identifier); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Identifier) Type() string {
	return "identifier"
}

func (s *Identifier) Literal() string {
	return s.Value
}

func (s *Identifier) String() string {
	return fmt.Sprintf("[ident: %s]", s.Value)
}

type Builtin struct {
	Name     string
	Token    *tokenizer.Token
	Args     []Expression
	Location helpers.Location
}

func (s *Builtin) IsEqual(node Node) bool {
	if node, ok := node.(*Builtin); ok {
		return s.Name == node.Name
	}
	return false
}

func (s *Builtin) Type() string {
	return "builtin"
}

func (s *Builtin) Literal() string {
	args := ""
	for _, arg := range s.Args {
		args += arg.Literal() + ", "
	}
	return fmt.Sprintf("%s(%s)", s.Name, args)
}

func (s *Builtin) String() string {
	args := ""
	for _, arg := range s.Args {
		args += arg.String() + ", "
	}
	return fmt.Sprintf("[builtin: %s(%s)]", s.Name, args)
}

type Boolean struct {
	Value bool
	Token *tokenizer.Token
}

func (s *Boolean) IsEqual(node Node) bool {
	if node, ok := node.(*Boolean); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Boolean) Type() string {
	return "boolean"
}

func (s *Boolean) String() string {
	return fmt.Sprintf("[bool: %v]", s.Value)
}

func (s *Boolean) Literal() string {
	return fmt.Sprintf("%v", s.Value)
}

type Integer struct {
	Value int
	Token *tokenizer.Token
}

func (s *Integer) IsEqual(node Node) bool {
	if node, ok := node.(*Integer); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Integer) Type() string {
	return "integer"
}

func (s *Integer) String() string {
	return fmt.Sprintf("[integer: %v]", s.Value)
}

func (s *Integer) Literal() string {
	return s.Token.Literal
}

type Float struct {
	Value float64
	Token *tokenizer.Token
}

func (s *Float) IsEqual(node Node) bool {
	if node, ok := node.(*Float); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Float) Type() string {
	return "float"
}

func (s *Float) String() string {
	return fmt.Sprintf("[float: %v]", s.Value)
}

func (s *Float) Literal() string {
	return s.Token.Literal
}

type Array struct {
	Items []Expression
	Token *tokenizer.Token
}

func (s *Array) IsEqual(node Node) bool {
	if node, ok := node.(*Array); ok {
		if len(s.Items) != len(node.Items) {
			return false
		}
		for i, item := range s.Items {
			if !item.IsEqual(node.Items[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (s *Array) Type() string {
	return "array"
}

func (s *Array) Literal() string {
	itemLiterals := []string{}
	for _, item := range s.Items {
		itemLiterals = append(itemLiterals, item.Literal())
	}
	items := strings.Join(itemLiterals, ", ")
	return fmt.Sprintf("[%s]", items)
}

func (s *Array) String() string {
	items := ""
	for _, item := range s.Items {
		items += item.String() + ", "
	}
	return fmt.Sprintf("[array: %s]", items)
}

type PrefixExpression struct {
	Token  *tokenizer.Token
	Op     string
	Action string

	Right Expression
}

func (s *PrefixExpression) IsEqual(node Node) bool {
	if node, ok := node.(*PrefixExpression); ok {
		return s.Op == node.Op && s.Right.IsEqual(node.Right)
	}
	return false
}

func (s *PrefixExpression) Type() string {
	return "prefix"
}

func (s *PrefixExpression) Literal() string {
	return fmt.Sprintf("%s%s", s.Op, s.Right.Literal())
}

func (s *PrefixExpression) String() string {
	return fmt.Sprintf("[prefix: %s%s]", s.Op, s.Right.String())
}

type InfixExpression struct {
	Token *tokenizer.Token
	Op    string

	Left  Expression
	Right Expression
}

func (s *InfixExpression) IsEqual(node Node) bool {
	if node, ok := node.(*InfixExpression); ok {
		return s.Op == node.Op && ((s.Left.IsEqual(node.Left) && s.Right.IsEqual(node.Right)) || (s.Left.IsEqual(node.Right) && s.Right.IsEqual(node.Left)))
	}
	return false
}

func (s *InfixExpression) Type() string {
	return "infix"
}

func (s *InfixExpression) Literal() string {
	return fmt.Sprintf("(%s %s %s)", s.Left.Literal(), s.Op, s.Right.Literal())
}

func (s *InfixExpression) String() string {
	return fmt.Sprintf("[infix: %s %s %s]", s.Left.String(), s.Op, s.Right.String())
}

type StringLiteral struct {
	Value string
	Token *tokenizer.Token
}

func (s *StringLiteral) IsEqual(node Node) bool {
	if node, ok := node.(*StringLiteral); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *StringLiteral) Type() string {
	return "string"
}

func (s *StringLiteral) Literal() string {
	return fmt.Sprintf(`"%s"`, s.Value)
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("[string: \"%s\"]", s.Value)
}

type FunctionCall struct {
	Called Expression
	Args   []Expression
	Token  *tokenizer.Token
}

func (s *FunctionCall) IsEqual(node Node) bool {
	if node, ok := node.(*FunctionCall); ok {
		if !s.Called.IsEqual(node.Called) {
			return false
		}
		if len(s.Args) != len(node.Args) {
			return false
		}
		for i, arg := range s.Args {
			if !arg.IsEqual(node.Args[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (s *FunctionCall) String() string {
	args := ""
	for _, arg := range s.Args {
		args += arg.String() + ", "
	}
	return fmt.Sprintf("[function: %s(%s)]", s.Called, args)
}

func (s *FunctionCall) Type() string {
	return "function"
}

func (s *FunctionCall) Literal() string {
	args := ""
	for _, arg := range s.Args {
		args += arg.Literal() + ", "
	}
	return fmt.Sprintf("%s(%s)", s.Called, args)
}

type ArrayAccess struct {
	Accessed Expression
	Index    Expression
	Token    *tokenizer.Token
}

func (s *ArrayAccess) IsEqual(node Node) bool {
	if node, ok := node.(*ArrayAccess); ok {
		return s.Accessed.IsEqual(node.Accessed) && s.Index.IsEqual(node.Index)
	}
	return false
}

func (s *ArrayAccess) Type() string {
	return "array_access"
}

func (s *ArrayAccess) Literal() string {
	return fmt.Sprintf("%s[%s]", s.Accessed.Literal(), s.Index.Literal())
}

func (s *ArrayAccess) String() string {
	return fmt.Sprintf("[array_access: %s[%s]]", s.Accessed.String(), s.Index.String())
}

type VariableAccess struct {
	Token      *tokenizer.Token
	Left       Expression
	IsOptional bool
	Right      *Identifier
}

func (s *VariableAccess) IsEqual(node Node) bool {
	if node, ok := node.(*VariableAccess); ok {
		return s.Left.IsEqual(node.Left) && s.Right.IsEqual(node.Right) && s.IsOptional == node.IsOptional
	}
	return false
}

func (s *VariableAccess) Type() string {
	return "variable"
}

func (s *VariableAccess) Literal() string {
	accessor := "."
	if s.IsOptional {
		accessor = "?."
	}
	return fmt.Sprintf("%s%s%s", s.Left.Literal(), accessor, s.Right.Literal())
}

func (s *VariableAccess) String() string {
	accessor := "."
	if s.IsOptional {
		accessor = "?."
	}
	return fmt.Sprintf("[variable: %s%s%s]", s.Left.String(), accessor, s.Right.String())
}

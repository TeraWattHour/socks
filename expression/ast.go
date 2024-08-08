package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
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
	Kind() string
	Location() helpers.Location
}

type DotAccess struct {
	Token    Token
	Property string
}

func (s *DotAccess) Location() helpers.Location {
	return s.Token.Location
}

func (s *DotAccess) IsEqual(node Node) bool {
	if dot, ok := node.(*DotAccess); ok {
		return dot.Property == s.Property
	}
	return false
}

func (s *DotAccess) Kind() string {
	return "dot_access"
}

func (s *DotAccess) Literal() string {
	return "." + s.Property
}

func (s *DotAccess) String() string {
	return fmt.Sprintf("[dot: %s]", s.Property)
}

type OptionalAccess struct {
	Token Token
}

func (s *OptionalAccess) Location() helpers.Location {
	return s.Token.Location
}

func (s *OptionalAccess) IsEqual(node Node) bool {
	_, ok := node.(*OptionalAccess)
	return ok
}

func (s *OptionalAccess) Kind() string {
	return "optional_access"
}

func (s *OptionalAccess) Literal() string {
	return "?."
}

func (s *OptionalAccess) String() string {
	return "[optional]"
}

type Identifier struct {
	Value string
	Token Token
}

func (s *Identifier) Location() helpers.Location {
	return s.Token.Location
}

func (s *Identifier) IsEqual(node Node) bool {
	if node, ok := node.(*Identifier); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Identifier) Kind() string {
	return "identifier"
}

func (s *Identifier) Literal() string {
	return s.Value
}

func (s *Identifier) String() string {
	return fmt.Sprintf("[id: %s]", s.Value)
}

type Nil struct {
	Token Token
}

func (s *Nil) Location() helpers.Location {
	return s.Token.Location
}

func (s *Nil) IsEqual(node Node) bool {
	_, ok := node.(*Nil)
	return ok
}

func (s *Nil) Kind() string {
	return "nil"
}

func (s *Nil) Literal() string {
	return "nil"
}

func (s *Nil) String() string {
	return "[nil]"
}

type Boolean struct {
	Value bool
	Token Token
}

func (s *Boolean) Location() helpers.Location {
	return s.Token.Location
}

func (s *Boolean) IsEqual(node Node) bool {
	if node, ok := node.(*Boolean); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Boolean) Kind() string {
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
	Token Token
}

func (s *Integer) Location() helpers.Location {
	return s.Token.Location
}

func (s *Integer) IsEqual(node Node) bool {
	if node, ok := node.(*Integer); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Integer) Kind() string {
	return "integer"
}

func (s *Integer) String() string {
	return fmt.Sprintf("[int: %v]", s.Value)
}

func (s *Integer) Literal() string {
	return s.Token.Literal
}

type Float struct {
	Value float64
	Token Token
}

func (s *Float) Location() helpers.Location {
	return s.Token.Location
}

func (s *Float) IsEqual(node Node) bool {
	if node, ok := node.(*Float); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *Float) Kind() string {
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
	Token Token
}

func (s *Array) Location() helpers.Location {
	return s.Token.Location
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

func (s *Array) Kind() string {
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
	Token  Token
	Op     TokenKind
	Action string

	Right Expression
}

func (s *PrefixExpression) Location() helpers.Location {
	return s.Token.Location
}

func (s *PrefixExpression) IsEqual(node Node) bool {
	if node, ok := node.(*PrefixExpression); ok {
		return s.Op == node.Op && s.Right.IsEqual(node.Right)
	}
	return false
}

func (s *PrefixExpression) Kind() string {
	return "prefix"
}

func (s *PrefixExpression) Literal() string {
	return fmt.Sprintf("%s%s", s.Op, s.Right.Literal())
}

func (s *PrefixExpression) String() string {
	return fmt.Sprintf("[prefix: %s%s]", s.Op, s.Right.String())
}

type InfixExpression struct {
	Token Token
	Op    TokenKind

	Left  Expression
	Right Expression
}

func (s *InfixExpression) Location() helpers.Location {
	return s.Token.Location
}

func (s *InfixExpression) IsEqual(node Node) bool {
	if node, ok := node.(*InfixExpression); ok {
		return s.Op == node.Op && ((s.Left.IsEqual(node.Left) && s.Right.IsEqual(node.Right)) || (s.Left.IsEqual(node.Right) && s.Right.IsEqual(node.Left)))
	}
	return false
}

func (s *InfixExpression) Kind() string {
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
	Token Token
}

func (s *StringLiteral) Location() helpers.Location {
	return s.Token.Location
}

func (s *StringLiteral) IsEqual(node Node) bool {
	if node, ok := node.(*StringLiteral); ok {
		return s.Value == node.Value
	}
	return false
}

func (s *StringLiteral) Kind() string {
	return "string"
}

func (s *StringLiteral) Literal() string {
	return fmt.Sprintf(`"%s"`, s.Value)
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("[string: \"%s\"]", s.Value)
}

type FunctionCall struct {
	Args       []Expression
	Token      Token
	closeToken Token
}

func (s *FunctionCall) Location() helpers.Location {
	return s.Token.Location
}

func (s *FunctionCall) IsEqual(node Node) bool {
	if node, ok := node.(*FunctionCall); ok {
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
	return fmt.Sprintf("[function: (%s)]", args)
}

func (s *FunctionCall) Kind() string {
	return "function"
}

func (s *FunctionCall) Literal() string {
	args := ""
	for _, arg := range s.Args {
		args += arg.Literal() + ", "
	}
	return fmt.Sprintf("(%s)", args)
}

type FieldAccess struct {
	Index      Expression
	Token      Token
	closeToken Token
}

func (s *FieldAccess) Location() helpers.Location {
	return s.Token.Location
}

func (s *FieldAccess) IsEqual(node Node) bool {
	if node, ok := node.(*FieldAccess); ok {
		return s.Index.IsEqual(node.Index)
	}
	return false
}

func (s *FieldAccess) Kind() string {
	return "fieldAccess"
}

func (s *FieldAccess) Literal() string {
	return fmt.Sprintf("[%s]", s.Index.Literal())
}

func (s *FieldAccess) String() string {
	return fmt.Sprintf("[fieldAccess: [%s]]", s.Index.String())
}

type Chain struct {
	Token Token
	Parts helpers.Queue[Expression]
}

func (s *Chain) Location() helpers.Location {
	return s.Token.Location
}

func (s *Chain) IsEqual(node Node) bool {
	if node, ok := node.(*Chain); ok {
		if len(s.Parts) != len(node.Parts) {
			return false
		}
		for i, part := range s.Parts {
			if !part.IsEqual(node.Parts[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (s *Chain) Kind() string {
	return "variable"
}

func (s *Chain) Literal() string {
	stringified := ""
	for _, part := range s.Parts {
		stringified += part.Literal()
	}

	return stringified
}

func (s *Chain) String() string {
	stringified := ""
	for _, part := range s.Parts {
		stringified += part.String()
	}

	return fmt.Sprintf("[chain: %s]", stringified)
}

type Ternary struct {
	Token       Token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (s *Ternary) Location() helpers.Location {
	return s.Token.Location
}

func (s *Ternary) IsEqual(node Node) bool {
	if node, ok := node.(*Ternary); ok {
		return s.Condition.IsEqual(node.Condition) && s.Consequence.IsEqual(node.Consequence) && s.Alternative.IsEqual(node.Alternative)
	}
	return false
}

func (s *Ternary) Kind() string {
	return "ternary"
}

func (s *Ternary) Literal() string {
	return fmt.Sprintf("%s ? %s : %s", s.Condition.Literal(), s.Consequence.Literal(), s.Alternative.Literal())
}

func (s *Ternary) String() string {
	return fmt.Sprintf("[ternary: %s ? %s : %s]", s.Condition.String(), s.Consequence.String(), s.Alternative.String())
}

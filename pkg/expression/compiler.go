package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/errors"
	"strings"
)

type Chunk struct {
	Instructions []int
	Constants    []any
	Lookups      []Expression
}

type Compiler struct {
	expr  Expression
	chunk Chunk
}

func NewCompiler(expr Expression) *Compiler {
	return &Compiler{
		expr: expr,
		chunk: Chunk{
			Instructions: make([]int, 0),
			Constants:    make([]any, 0),
			Lookups:      make([]Expression, 0),
		},
	}
}

func (c *Compiler) Compile() (Chunk, error) {
	if err := c.compile(c.expr); err != nil {
		return Chunk{}, err
	}
	return c.chunk, nil
}

func (c *Compiler) compile(expr Expression) error {
	switch expr := expr.(type) {
	case *Array:
		for _, item := range expr.Items {
			if err := c.compile(item); err != nil {
				return err
			}
		}
		c.emit(OpArray)
		c.addLookup(expr)
		c.emit(len(expr.Items))
	case *Boolean:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
	case *Float:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
	case *Integer:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
	case *StringLiteral:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
	case *Identifier:
		c.emit(OpGet)
		c.addLookup(expr)
		c.emit(c.addConstant(expr.Value))
	case *Builtin:
		for _, arg := range expr.Args {
			if err := c.compile(arg); err != nil {
				return err
			}
		}

		if expr.Name == "range" && len(expr.Args) > 1 {
			if len(expr.Args) == 2 {
				c.emitConstant(1)
			}
			c.emit(OpBuiltin3)
			c.addLookup(expr)
			c.emit(builtinRelativeIndex(expr.Name))
			break
		}

		builtinType := builtinType(expr.Name)

		if builtinType == -1 {
			panic("unknown builtin: " + expr.Name)
		}

		if len(expr.Args) != builtinType {
			types := builtinTypes[expr.Name]
			inputTypes := types[:len(types)-1]
			returnType := types[len(types)-1]
			return errors.New(fmt.Sprintf("call to %s(%s) -> any does not match the signature of %s(%s) -> %s", expr.Name, strings.TrimSuffix(strings.Repeat("any, ", len(expr.Args)), ", "), expr.Name, strings.Join(inputTypes, ", "), returnType), expr.Location)
		}

		switch builtinType {
		case 1:
			c.emit(OpBuiltin1)
		case 2:
			c.emit(OpBuiltin2)
		case 3:
			c.emit(OpBuiltin3)
		}

		c.addLookup(expr)
		c.emit(builtinRelativeIndex(expr.Name))
	case *FunctionCall:
		if err := c.compile(expr.Called); err != nil {
			return err
		}
		for _, arg := range expr.Args {
			if err := c.compile(arg); err != nil {
				return err
			}
		}
		c.emit(OpCall)
		c.addLookup(expr)
		c.emit(len(expr.Args))
	case *ArrayAccess:
		if err := c.compile(expr.Accessed); err != nil {
			return err
		}
		if err := c.compile(expr.Index); err != nil {
			return err
		}
		c.emit(OpArrayAccess)
		c.addLookup(expr)
	case *VariableAccess:
		if err := c.compile(expr.Left); err != nil {
			return err
		}
		if expr.IsOptional {
			c.emit(OpOptionalChain)
		} else {
			c.emit(OpChain)
		}
		c.addLookup(expr)
		c.emit(c.addConstant(expr.Right.Value))
	case *InfixExpression:
		if err := c.compile(expr.Left); err != nil {
			return err
		}
		if err := c.compile(expr.Right); err != nil {
			return err
		}
		switch expr.Op {
		case "and":
			c.emit(OpAnd)
		case "or":
			c.emit(OpOr)
		case "eq":
			c.emit(OpEq)
		case "neq":
			c.emit(OpNeq)
		case "lt":
			c.emit(OpLt)
		case "lte":
			c.emit(OpLte)
		case "gt":
			c.emit(OpGt)
		case "gte":
			c.emit(OpGte)
		case "plus":
			c.emit(OpAdd)
		case "minus":
			c.emit(OpSubtract)
		case "multiply":
			c.emit(OpMultiply)
		case "divide":
			c.emit(OpDivide)
		case "in":
			c.emit(OpIn)
		case "exponent":
			c.emit(OpExponent)
		case "modulo":
			c.emit(OpModulo)
		}
		c.addLookup(expr)
	case *PrefixExpression:
		if err := c.compile(expr.Right); err != nil {
			return err
		}
		switch expr.Op {
		case "not":
			c.emit(OpNot)
		case "minus":
			c.emit(OpNegate)
		}
		c.addLookup(expr)
	}
	return nil
}

func (c *Compiler) emit(op int) {
	c.chunk.Instructions = append(c.chunk.Instructions, op)
}

func (c *Compiler) emitConstant(value any) {
	c.chunk.Constants = append(c.chunk.Constants, value)
	c.emit(OpConstant)
	c.emit(len(c.chunk.Constants) - 1)
}

func (c *Compiler) addConstant(value any) int {
	c.chunk.Constants = append(c.chunk.Constants, value)
	return len(c.chunk.Constants) - 1
}

func (c *Compiler) addLookup(expression Expression) {
	atIndex := len(c.chunk.Instructions) - 1
	fillSpaces := atIndex - len(c.chunk.Lookups) + 1
	if fillSpaces > 0 {
		c.chunk.Lookups = append(c.chunk.Lookups, make([]Expression, fillSpaces)...)
	}
	c.chunk.Lookups[atIndex] = expression
}

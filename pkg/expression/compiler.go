package expression

import (
	"fmt"
	"slices"
)

type Chunk struct {
	Instructions []int
	Constants    []any
}

func (c *Chunk) String() string {
	return fmt.Sprintf("Instructions: %v\nConstants: %v", c.Instructions, c.Constants)
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
		},
	}
}

func (c *Compiler) Compile() (Chunk, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	c.compile(c.expr)
	return c.chunk, nil
}

func (c *Compiler) compile(expr Expression) {
	switch expr := expr.(type) {
	case *Array:
		for _, item := range expr.Items {
			c.compile(item)
		}
		c.emit(OpArray)
		c.emit(len(expr.Items))
	case *Boolean:
		c.emitConstant(expr.Value)
	case *Numeric:
		c.emitConstant(expr.Value)
	case *Integer:
		c.emitConstant(expr.Value)
	case *StringLiteral:
		c.emitConstant(expr.Value)
	case *Identifier:
		c.emit(OpGet)
		c.emit(c.addConstant(expr.Value))
	case *Builtin:
		for _, arg := range expr.Args {
			c.compile(arg)
		}
		builtinType := builtinType(expr.Name)
		if len(expr.Args) != builtinType {
			panic("wrong number of arguments provided")
		}
		if builtinType == 1 {
			c.emit(OpBuiltin1)
			c.emit(slices.Index(builtinNames, expr.Name))
		} else if builtinType == 2 {
			c.emit(OpBuiltin2)
			c.emit(slices.Index(builtinNames, expr.Name))
		} else if builtinType == 3 {
			c.emit(OpBuiltin3)
			c.emit(slices.Index(builtinNames, expr.Name))
		}
	case *FunctionCall:
		c.compile(expr.Called)
		for _, arg := range expr.Args {
			c.compile(arg)
		}
		c.emit(OpCall)
		c.emit(len(expr.Args))
	case *ArrayAccess:
		c.compile(expr.Accessed)
		c.compile(expr.Index)
		c.emit(OpArrayAccess)
	case *VariableAccess:
		c.compile(expr.Left)
		c.emit(OpChain)
		c.compile(expr.Right)
	case *InfixExpression:
		c.compile(expr.Left)
		c.compile(expr.Right)
		switch expr.Op {
		case "and":
			c.emit(OpAnd)
		case "or":
			c.emit(OpOr)
		case "eq":
			c.emit(OpEqual)
		case "neq":
			c.emit(OpNotEqual)
		case "lt":
			c.emit(OpLessThan)
		case "lte":
			c.emit(OpLessThanOrEqual)
		case "gt":
			c.emit(OpGreaterThan)
		case "gte":
			c.emit(OpGreaterThanOrEqual)
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
		}
	case *PrefixExpression:
		c.compile(expr.Right)
		switch expr.Op {
		case "not":
			c.emit(OpNot)
		case "minus":
			c.emit(OpNegate)
		}
	}
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

package expression

import (
	"fmt"
	"reflect"
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

var builtinNames = []string{
	"float32",
	"float64",
	"int",
	"int8",
	"int16",
	"int32",
	"int64",
	"uint",
	"uint8",
	"uint16",
	"uint32",
	"uint64",
	"uintptr",
	"len",
	"range",
	"rangeStep",
}

var builtinsOne = map[string]func(val any) any{
	"float32": castFloat32,
	"float64": castFloat64,
	"int":     castInt,
	"int8":    castInt8,
	"int16":   castInt16,
	"int32":   castInt32,
	"int64":   castInt64,
	"uint":    castUint,
	"uint8":   castUint8,
	"uint16":  castUint16,
	"uint32":  castUint32,
	"uint64":  castUint64,
	"uintptr": castUintptr,
	"len":     length,
}

var builtinsTwo = map[string]func(val1, val2 any) any{
	"range": rangeArray,
}

var builtinsThree = map[string]func(val1, val2, val3 any) any{
	"rangeStep": rangeArrayStep,
}

var numBuiltinsOne = reflect.ValueOf(builtinsOne).Len()
var numBuiltinsTwo = reflect.ValueOf(builtinsTwo).Len()
var numBuiltinsThree = reflect.ValueOf(builtinsThree).Len()

func builtinType(name string) int {
	idx := slices.Index(builtinNames, name)
	if idx == -1 {
		panic("not a builtin")
	} else if idx < numBuiltinsOne {
		return 1
	} else if idx < numBuiltinsOne+numBuiltinsTwo {
		return 2
	} else if idx < numBuiltinsOne+numBuiltinsTwo+numBuiltinsThree {
		return 3
	}

	panic("not implemented")
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
			fmt.Println("here")
			fmt.Println(err)
		}
	}()
	c.compile(c.expr, false)
	return c.chunk, nil
}

func (c *Compiler) compile(expr Expression, chain bool) {
	switch expr := expr.(type) {
	case *Array:
		for _, item := range expr.Items {
			c.compile(item, false)
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
			c.compile(arg, false)
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
		c.compile(expr.Called, chain)
		for _, arg := range expr.Args {
			c.compile(arg, false)
		}
		c.emit(OpCall)
		c.emit(len(expr.Args))
	case *ArrayAccess:
		c.compile(expr.Accessed, chain)
		c.compile(expr.Index, false)
		c.emit(OpArrayAccess)
	case *VariableAccess:
		c.compile(expr.Left, chain)
		c.emit(OpChain)
		c.compile(expr.Right, true)
	case *InfixExpression:
		c.compile(expr.Left, chain)
		c.compile(expr.Right, chain)
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
		c.compile(expr.Right, chain)
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

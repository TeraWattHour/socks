package expression

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
	"strings"
)

type Chunk struct {
	Instructions []int
	Constants    []any
	Lookups      []Expression
}

type Compiler struct {
	expr           Expression
	chunk          Chunk
	optionalChains map[Expression][]int
}

func NewCompiler(expr Expression) *Compiler {
	return &Compiler{
		expr:           expr,
		optionalChains: map[Expression][]int{},
		chunk: Chunk{
			Instructions: make([]int, 0),
			Constants:    make([]any, 0),
			Lookups:      make([]Expression, 0),
		},
	}
}

func (c *Compiler) Compile() (Chunk, error) {
	if err := c.compile(c.expr, c.expr); err != nil {
		return Chunk{}, err
	}
	return c.chunk, nil
}

func (c *Compiler) compile(expr Expression, scope Expression) error {
	switch expr := expr.(type) {
	case *Array:
		for _, item := range expr.Items {
			if err := c.compile(item, item); err != nil {
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
		c.emit(c.createConstant(expr.Value))
	case *Nil:
		c.emit(OpNil)
		c.addLookup(expr)
	case *Builtin:
		for _, arg := range expr.Args {
			if err := c.compile(arg, arg); err != nil {
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
			return errors.New_(fmt.Sprintf("call to %s(%s) -> any does not match the signature of %s(%s) -> %s", expr.Name, strings.TrimSuffix(strings.Repeat("any, ", len(expr.Args)), ", "), expr.Name, strings.Join(inputTypes, ", "), returnType), expr.location)
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
		if err := c.compile(expr.Called, scope); err != nil {
			return err
		}
		for _, arg := range expr.Args {
			if err := c.compile(arg, arg); err != nil {
				return err
			}
		}
		c.emit(OpCall)
		c.addLookup(expr)
		c.emit(len(expr.Args))
	case *FieldAccess:
		err := c.compile(expr.Accessed, scope)
		if err != nil {
			return err
		}

		if err := c.compile(expr.Index, expr.Index); err != nil {
			return err
		}

		c.emit(OpArrayAccess)
		c.addLookup(expr)
	case *Chain:
		if err := c.compile(expr.Left, scope); err != nil {
			return err
		}
		if expr.IsOptional {
			c.emit(OpOptionalChain)
			c.addLookup(expr)
			c.emit(-1)
			c.optionalChains[scope] = append(c.optionalChains[scope], len(c.chunk.Instructions)-1)
		} else {
			c.emit(OpChain)
			c.addLookup(expr)
		}
		c.emit(c.createConstant(expr.Right.Value))
	case *Ternary:
		if err := c.compile(expr.Condition, expr.Condition); err != nil {
			return err
		}
		c.emit(OpTernary)
		start := len(c.chunk.Instructions)
		c.emit(-1)
		if err := c.compile(expr.Consequence, expr.Consequence); err != nil {
			return err
		}
		c.emit(OpJmp)
		c.emit(-1)
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
		start = len(c.chunk.Instructions) - 1
		if err := c.compile(expr.Alternative, expr.Alternative); err != nil {
			return err
		}
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
	case *InfixExpression:
		var err error
		if err = c.compile(expr.Left, expr.Left); err != nil {
			return err
		}
		if expr.Op != tokenizer.TokElvis {
			if err = c.compile(expr.Right, expr.Right); err != nil {
				return err
			}
		}

		switch expr.Op {
		case tokenizer.TokAnd:
			c.emit(OpAnd)
		case tokenizer.TokOr:
			c.emit(OpOr)
		case tokenizer.TokEq:
			c.emit(OpEq)
		case tokenizer.TokNeq:
			c.emit(OpNeq)
		case tokenizer.TokLt:
			c.emit(OpLt)
		case tokenizer.TokLte:
			c.emit(OpLte)
		case tokenizer.TokGt:
			c.emit(OpGt)
		case tokenizer.TokGte:
			c.emit(OpGte)
		case tokenizer.TokPlus:
			c.emit(OpAdd)
		case tokenizer.TokMinus:
			c.emit(OpSubtract)
		case tokenizer.TokAsterisk:
			c.emit(OpMultiply)
		case tokenizer.TokSlash:
			c.emit(OpDivide)
		case tokenizer.TokIn:
			c.emit(OpIn)
		case tokenizer.TokPower:
			c.emit(OpPower)
		case tokenizer.TokModulo:
			c.emit(OpModulo)
		case tokenizer.TokElvis:
			c.emit(OpElvis)
			start := len(c.chunk.Instructions)
			c.emit(-1)
			if err = c.compile(expr.Right, expr.Right); err != nil {
				return err
			}
			c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
		}
		c.addLookup(expr)
	case *PrefixExpression:
		if err := c.compile(expr.Right, expr.Right); err != nil {
			return err
		}

		switch expr.Op {
		case tokenizer.TokNot:
			c.emit(OpNot)
		case tokenizer.TokMinus:
			c.emit(OpNegate)
		}
		c.addLookup(expr)
	}

	if expr == scope {
		c.updateChainJumps(scope)
	}

	return nil
}

func (c *Compiler) updateChainJumps(expr Expression) {
	if c.optionalChains[expr] == nil {
		return
	}
	for _, ip := range c.optionalChains[expr] {
		c.chunk.Instructions[ip] = len(c.chunk.Instructions)
	}
	delete(c.optionalChains, expr)
}

func (c *Compiler) emit(op int) {
	c.chunk.Instructions = append(c.chunk.Instructions, op)
}

func (c *Compiler) createConstant(value any) int {
	found := slices.Index(c.chunk.Constants, value)
	if found != -1 {
		return found
	}
	c.chunk.Constants = append(c.chunk.Constants, value)
	return len(c.chunk.Constants) - 1
}

func (c *Compiler) emitConstant(value any) {
	c.emit(OpConstant)
	c.emit(c.createConstant(value))
}

func (c *Compiler) addLookup(expression Expression) {
	atIndex := len(c.chunk.Instructions) - 1
	fillSpaces := atIndex - len(c.chunk.Lookups) + 1
	if fillSpaces > 0 {
		c.chunk.Lookups = append(c.chunk.Lookups, make([]Expression, fillSpaces)...)
	}
	c.chunk.Lookups[atIndex] = expression
}

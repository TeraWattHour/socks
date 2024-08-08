package expression

import (
	"github.com/terawatthour/socks/errors"
	"slices"
)

type Program struct {
	Instructions []int
	Constants    []any
	Lookups      []Expression
}

type Compiler struct {
	expr  Expression
	chunk Program
}

func NewCompiler(expr Expression) *Compiler {
	return &Compiler{
		expr: expr,
		chunk: Program{
			Instructions: make([]int, 0),
			Constants:    make([]any, 0),
			Lookups:      make([]Expression, 0),
		},
	}
}

func (c *Compiler) Compile() (Program, error) {
	if err := c.compile(c.expr); err != nil {
		return Program{}, err
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
		c.emit(c.createConstant(expr.Value))
	case *Nil:
		c.emit(OpNil)
		c.addLookup(expr)
	case *Chain:
		var optionalIndices []int

		for i, part := range expr.Parts {
			switch part := part.(type) {
			case *FieldAccess:
				if err := c.compile(part.Index); err != nil {
					return err
				}
				c.emit(OpPropertyAccess)
				c.addLookup(part)
			case *DotAccess:
				c.emit(OpChain)
				c.addLookup(part)
				c.emit(c.createConstant(part.Property))
			case *OptionalAccess:
				c.emit(OpOptionalChain)
				c.addLookup(part)
				optionalIndices = append(optionalIndices, len(c.chunk.Instructions))
				c.emit(-1)
			case *FunctionCall:
				for _, arg := range part.Args {
					if err := c.compile(arg); err != nil {
						return err
					}
				}

				c.emit(OpCall)
				c.addLookup(part)
				c.emit(len(part.Args))
			case *Identifier:
				if i == 0 {
					if err := c.compile(part); err != nil {
						return err
					}
				} else {
					c.emit(OpChain)
					c.addLookup(part)
					c.emit(c.createConstant(part.Value))
				}
			default:
				if err := c.compile(part); err != nil {
					return err
				}
			}
		}

		for _, index := range optionalIndices {
			c.chunk.Instructions[index] = len(c.chunk.Instructions) - index
		}
	case *Ternary:
		if err := c.compile(expr.Condition); err != nil {
			return err
		}
		c.emit(OpTernary)
		start := len(c.chunk.Instructions)
		c.emit(-1)
		if err := c.compile(expr.Consequence); err != nil {
			return err
		}
		c.emit(OpJmp)
		c.emit(-1)
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
		start = len(c.chunk.Instructions) - 1
		if err := c.compile(expr.Alternative); err != nil {
			return err
		}
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
	case *InfixExpression:
		var err error
		if err = c.compile(expr.Left); err != nil {
			return err
		}
		if expr.Op != TokElvis {
			if err = c.compile(expr.Right); err != nil {
				return err
			}
		}

		switch expr.Op {
		case TokAnd:
			c.emit(OpAnd)
		case TokOr:
			c.emit(OpOr)
		case TokEq:
			c.emit(OpEq)
		case TokNeq:
			c.emit(OpNeq)
		case TokLt:
			c.emit(OpLt)
		case TokLte:
			c.emit(OpLte)
		case TokGt:
			c.emit(OpGt)
		case TokGte:
			c.emit(OpGte)
		case TokPlus:
			c.emit(OpAdd)
		case TokMinus:
			c.emit(OpSubtract)
		case TokAsterisk:
			c.emit(OpMultiply)
		case TokSlash:
			c.emit(OpDivide)
		case TokIn:
			c.emit(OpIn)
		case TokPower:
			c.emit(OpPower)
		case TokModulo:
			c.emit(OpModulo)
		case TokElvis:
			c.emit(OpElvis)
			start := len(c.chunk.Instructions)
			c.emit(-1)
			if err = c.compile(expr.Right); err != nil {
				return err
			}
			c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
		}
		c.addLookup(expr)
	case *PrefixExpression:
		if err := c.compile(expr.Right); err != nil {
			return err
		}

		switch expr.Op {
		case TokNot, TokBang:
			c.emit(OpNot)
		case TokMinus:
			c.emit(OpNegate)
		}
		c.addLookup(expr)
	}

	return nil
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

func (c *Compiler) error(message string, token Token) error {
	return errors.New(message, token.Location)
}

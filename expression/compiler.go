package expression

import (
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
)

type Chunk struct {
	Instructions []int
	Constants    []any
	Lookups      []Expression
}

type Compiler struct {
	file  helpers.File
	expr  Expression
	chunk Chunk
}

func NewCompiler(file helpers.File, expr Expression) *Compiler {
	return &Compiler{
		file: file,
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
			c.chunk.Instructions[index] = len(c.chunk.Instructions) - 1
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
		if expr.Op != tokenizer.TokElvis {
			if err = c.compile(expr.Right); err != nil {
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
		case tokenizer.TokNot, tokenizer.TokBang:
			c.emit(OpNot)
		case tokenizer.TokMinus:
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

func (c *Compiler) error(message string, location helpers.Location) error {
	return errors.New(message, c.file.Name, c.file.Content, location, location.FromOther())
}

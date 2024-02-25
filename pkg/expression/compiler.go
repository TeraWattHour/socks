package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"slices"
	"strings"
)

type Type string

const (
	TInt     Type = "int"
	TBool    Type = "bool"
	TFloat   Type = "float"
	TString  Type = "string"
	TArray   Type = "array"
	TNil     Type = "nil"
	TUnknown Type = "unknown"
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
	if err, _ := c.compile(c.expr, c.expr); err != nil {
		return Chunk{}, err
	}
	return c.chunk, nil
}

func (c *Compiler) compile(expr Expression, scope Expression) (error, Type) {
	var returnedType Type = TUnknown

	switch expr := expr.(type) {
	case *Array:
		for _, item := range expr.Items {
			if err, _ := c.compile(item, item); err != nil {
				return err, ""
			}
		}
		c.emit(OpArray)
		c.addLookup(expr)
		c.emit(len(expr.Items))
		returnedType = TArray
	case *Boolean:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
		returnedType = TBool
	case *Float:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
		returnedType = TFloat
	case *Integer:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
		returnedType = TInt
	case *StringLiteral:
		c.emitConstant(expr.Value)
		c.addLookup(expr)
		returnedType = TString
	case *Identifier:
		c.emit(OpGet)
		c.addLookup(expr)
		c.emit(c.createConstant(expr.Value))
		returnedType = TUnknown
	case *Nil:
		c.emit(OpNil)
		c.addLookup(expr)
		returnedType = TNil
	case *Builtin:
		var argTypes []Type

		for _, arg := range expr.Args {
			if err, returned := c.compile(arg, arg); err != nil {
				return err, ""
			} else {
				argTypes = append(argTypes, returned)
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
			return errors.New(fmt.Sprintf("call to %s(%s) -> any does not match the signature of %s(%s) -> %s", expr.Name, strings.TrimSuffix(strings.Repeat("any, ", len(expr.Args)), ", "), expr.Name, strings.Join(inputTypes, ", "), returnType), expr.location), ""
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
		if err, returned := c.compile(expr.Called, scope); err != nil {
			return err, ""
		} else if returned != TUnknown {
			return errors.New("cannot call a non-function", expr.Location()), ""
		}
		for _, arg := range expr.Args {
			if err, _ := c.compile(arg, arg); err != nil {
				return err, ""
			}
		}
		c.emit(OpCall)
		c.addLookup(expr)
		c.emit(len(expr.Args))
		returnedType = TUnknown
	case *FieldAccess:
		err, accessedType := c.compile(expr.Accessed, scope)
		if err != nil {
			return err, ""
		}

		if err, returned := c.compile(expr.Index, expr.Index); err != nil {
			return err, ""
		} else if returned != TInt && accessedType == TArray {
			return errors.New("slice access index must be an integer", expr.Index.Location()), ""
		}

		c.emit(OpArrayAccess)
		c.addLookup(expr)
		returnedType = TUnknown
	case *Chain:
		if err, _ := c.compile(expr.Left, scope); err != nil {
			return err, ""
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
		if err, _ := c.compile(expr.Condition, expr.Condition); err != nil {
			return err, ""
		}
		c.emit(OpTernary)
		start := len(c.chunk.Instructions)
		c.emit(-1)
		if err, _ := c.compile(expr.Consequence, expr.Consequence); err != nil {
			return err, ""
		}
		c.emit(OpJmp)
		c.emit(-1)
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
		start = len(c.chunk.Instructions) - 1
		if err, _ := c.compile(expr.Alternative, expr.Alternative); err != nil {
			return err, ""
		}
		c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
	case *InfixExpression:
		var err error
		var leftType, rightType Type
		if err, leftType = c.compile(expr.Left, expr.Left); err != nil {
			return err, ""
		}
		if expr.Op != tokenizer.TokElvis {
			if err, rightType = c.compile(expr.Right, expr.Right); err != nil {
				return err, ""
			}
		}

		known := leftType != TUnknown && rightType != TUnknown

		switch expr.Op {
		case tokenizer.TokEq, tokenizer.TokNeq:
			if known && leftType != rightType {
				return errors.New(fmt.Sprintf("cannot equate %s to %s", leftType, rightType), expr.Location()), ""
			}
		case tokenizer.TokLt, tokenizer.TokLte, tokenizer.TokGt, tokenizer.TokGte, tokenizer.TokMinus, tokenizer.TokAsterisk, tokenizer.TokSlash, tokenizer.TokPower:
			if known && !checkTypes(leftType, rightType, TInt, TFloat) {
				return errors.New(fmt.Sprintf("invalid operation, mismatched types: %v %v %v", leftType, expr.Token.Literal, rightType), expr.Location()), ""
			}
		case tokenizer.TokModulo:
			if known && !checkTypes(leftType, rightType, TInt) {
				return errors.New(fmt.Sprintf("invalid operation, mismatched types: %v %% %v", leftType, rightType), expr.Location()), ""
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
			if known && !checkTypes(leftType, rightType, TInt, TFloat) && !symmetricCheck(leftType, rightType, TString) {
				return errors.New(fmt.Sprintf("invalid operation, mismatched types: %v + %v", leftType, rightType), expr.Location()), ""
			}
			c.emit(OpAdd)
		case tokenizer.TokMinus:
			c.emit(OpSubtract)
		case tokenizer.TokAsterisk:
			c.emit(OpMultiply)
		case tokenizer.TokSlash:
			c.emit(OpDivide)
		case tokenizer.TokIn:
			if known && (rightType != TArray && rightType != TString) {
				return errors.New(fmt.Sprintf("cannot use 'in' with %s", rightType), expr.Location()), ""
			}
			c.emit(OpIn)
		case tokenizer.TokPower:
			c.emit(OpPower)
		case tokenizer.TokModulo:
			c.emit(OpModulo)
		case "elvis":
			c.emit(OpElvis)
			start := len(c.chunk.Instructions)
			c.emit(-1)
			if err, rightType = c.compile(expr.Right, expr.Right); err != nil {
				return err, ""
			}
			c.chunk.Instructions[start] = len(c.chunk.Instructions) - start
			returnedType = rightType
		}
		c.addLookup(expr)
	case *PrefixExpression:
		err, returned := c.compile(expr.Right, expr.Right)
		if err != nil {
			return err, ""
		}

		switch expr.Op {
		case tokenizer.TokNot:
			c.emit(OpNot)
			returnedType = TBool
		case tokenizer.TokMinus:
			c.emit(OpNegate)
			returnedType = returned

		}
		c.addLookup(expr)
	}

	if expr == scope {
		c.updateChainJumps(scope)
	}

	return nil, returnedType
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

func checkTypes(left Type, right Type, allowed ...Type) bool {
	return slices.Index(allowed, left) != -1 && slices.Index(allowed, right) != -1
}

func known(left, right Type) bool {
	return left != TUnknown && right != TUnknown
}

func symmetricCheck(left Type, right Type, allowed ...Type) bool {
	return !(known(left, right) && (left != right || !checkTypes(left, right, allowed...)))
}

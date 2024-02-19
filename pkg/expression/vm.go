package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"reflect"
)

type VM struct {
	chunk Chunk
	stack Stack
	funcs map[string]any
	ip    int
}

type Stack []any

func (s *Stack) push(v any) {
	*s = append(*s, v)
}

func (s *Stack) pop() any {
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

func NewVM(chunk Chunk) *VM {
	return &VM{
		chunk: chunk,
		stack: []any{},
	}
}

func (vm *VM) Run(env map[string]any) (any, error) {
outerLoop:
	for vm.ip = 0; vm.ip < len(vm.chunk.Instructions); vm.ip++ {
		var currentError error
		switch vm.chunk.Instructions[vm.ip] {
		case OpChain:
			left := vm.stack.pop()
			field := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.push(accessVariable(left, field))
			vm.ip++
		case OpOptionalChain:
			left := vm.stack.pop()
			field := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.push(accessVariable(left, field))

		case OpArrayAccess:
			_index := vm.stack.pop()
			_value := vm.stack.pop()
			val := reflect.ValueOf(_value)
			switch val.Kind() {
			case reflect.Array, reflect.Slice:
				if !reflect.ValueOf(_index).CanConvert(reflect.TypeOf(0)) {
					return nil, errors2.New(fmt.Sprintf("expected int, got %v", reflect.TypeOf(_index)), vm.chunk.Lookups[vm.ip].(*ArrayAccess).Token.LocationEnd)
				}
				index := reflect.ValueOf(_index).Convert(reflect.TypeOf(0)).Int()

				vm.stack.push(val.Index(int(index)).Interface())
			case reflect.Map:
				vm.stack.push(val.MapIndex(reflect.ValueOf(_index)).Interface())
			default:
				return nil, errors2.New(fmt.Sprintf("expected array or object, got %v", reflect.TypeOf(_value)), vm.chunk.Lookups[vm.ip].(*ArrayAccess).Token.LocationEnd)
			}
		case OpBuiltin1:
			fmt.Println("before call", vm.stack)

			function := builtinsOne[vm.chunk.Instructions[vm.ip+1]]
			arg := vm.stack.pop()
			result := function(arg)

			if general, ok := result.(error); ok {
				message := general.Error()
				return nil, errors2.New(message, vm.chunk.Lookups[vm.ip].(*Builtin).Location)
			}

			vm.stack.push(result)
			fmt.Println("after call", vm.stack)
			vm.ip++
		case OpBuiltin2:
			function := builtinsTwo[vm.chunk.Instructions[vm.ip+1]]
			arg2 := vm.stack.pop()
			arg1 := vm.stack.pop()
			result := function(arg1, arg2)
			if general, ok := result.(error); ok {
				message := general.Error()
				return nil, errors2.New(message, vm.chunk.Lookups[vm.ip].(*Builtin).Location)
			}
			vm.stack.push(result)
			vm.ip++
		case OpBuiltin3:
			function := builtinsThree[vm.chunk.Instructions[vm.ip+1]]
			arg3 := vm.stack.pop()
			arg2 := vm.stack.pop()
			arg1 := vm.stack.pop()
			result := function(arg1, arg2, arg3)
			if general, ok := result.(error); ok {
				message := general.Error()
				return nil, errors2.New(message, vm.chunk.Lookups[vm.ip].(*Builtin).Location)
			}
			vm.stack.push(result)
			vm.ip++
		case OpArray:
			count := vm.chunk.Instructions[vm.ip+1]
			items := make([]any, count)
			for j := 0; j < count; j++ {
				items[count-j-1] = vm.stack.pop()
			}
			vm.stack.push(items)
			vm.ip++

		case OpCall:
			argumentCount := vm.chunk.Instructions[vm.ip+1]
			vm.ip++
			args := make([]reflect.Value, argumentCount)
			for j := 0; j < argumentCount; j++ {
				args[j] = reflect.ValueOf(vm.stack.pop())
			}

			function := reflect.ValueOf(vm.stack.pop())

			results := function.Call(args)
			if len(results) == 1 {
				vm.stack.push(results[0].Interface())
			} else if len(results) > 1 {
				vm.stack.push(reflectedSliceToInterfaceSlice(results))
			}
		case OpGet:
			ident := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.push(env[ident])
			vm.ip++
		case OpConstant:
			constant := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]]
			vm.ip++
			vm.stack.push(constant)
		case OpIn:
			right := vm.stack.pop()
			left := vm.stack.pop()

			for i := 0; i < reflect.ValueOf(right).Len(); i++ {
				if reflect.ValueOf(right).Index(i).Interface() == left {
					vm.stack.push(true)
					continue outerLoop
				}
			}
			vm.stack.push(false)
		case OpEq:
			left := vm.stack.pop()
			right := vm.stack.pop()
			vm.stack.push(left == right)
		case OpNegate:
			vm.stack.push(negate(vm.stack.pop()))
		case OpNeq:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(left != right)
		case OpNot:
			vm.stack.push(!CastToBool(vm.stack.pop()))
		case OpAdd:
			currentError = vm.executeInfixExpression(binaryAddition)
		case OpLt:
			currentError = vm.executeInfixExpression(binaryLessThan)
		case OpGt:
			currentError = vm.executeInfixExpression(binaryGreaterThan)
		case OpGte:
			currentError = vm.executeInfixExpression(binaryGreaterThanEqual)
		case OpLte:
			currentError = vm.executeInfixExpression(binaryLessThanEqual)
		case OpSubtract:
			currentError = vm.executeInfixExpression(binarySubtraction)
		case OpMultiply:
			currentError = vm.executeInfixExpression(binaryMultiplication)
		case OpDivide:
			currentError = vm.executeInfixExpression(binaryDivision)
		case OpModulo:
			currentError = vm.executeInfixExpression(binaryModulo)
		case OpExponent:
			currentError = vm.executeInfixExpression(binaryExponentiation)
		case OpAnd:
			currentError = vm.executeInfixExpression(and)
		case OpOr:
			currentError = vm.executeInfixExpression(or)
		case OpCodeCount:
			panic("unexpected end of program")
		}
		if currentError != nil {
			return nil, currentError
		}
	}

	if len(vm.stack) == 0 {
		return nil, errors2.New("expression does not return a value", helpers.Location{-1, -1})
	}

	if len(vm.stack) != 1 {
		fmt.Println(vm.stack)
		return nil, fmt.Errorf("expression returns multiple values")
	}

	return vm.stack.pop(), nil
}

func (vm *VM) executeInfixExpression(fn func(any, any) any) error {
	right := vm.stack.pop()
	left := vm.stack.pop()
	res := fn(left, right)
	if general, ok := res.(error); ok {
		message := general.Error()
		return errors2.New(message, vm.chunk.Lookups[vm.ip].(*InfixExpression).Token.LocationStart)
	}
	vm.stack.push(res)

	return nil
}

func accessVariable(base any, field any) any {
	val := reflect.ValueOf(base)
	switch val.Kind() {
	case reflect.Map:
		return val.MapIndex(reflect.ValueOf(field)).Interface()
	case reflect.Struct:
		return val.FieldByName(field.(string)).Interface()
	default:
		panic(fmt.Sprintf("unsupported type for dot access, got %v", val.Kind()))
	}
}

func reflectedSliceToInterfaceSlice(vs []reflect.Value) []interface{} {
	is := make([]interface{}, len(vs))
	for i, v := range vs {
		is[i] = v.Interface()
	}
	return is
}

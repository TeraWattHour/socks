package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"reflect"
)

type VM struct {
	chunk        Chunk
	stack        helpers.Stack[any]
	funcs        map[string]any
	ip           int
	currentError error
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
		vm.currentError = nil
		switch vm.chunk.Instructions[vm.ip] {
		case OpChain:
			object := vm.stack.Pop()
			if object == nil {
				return nil, errors2.New("cannot access properties of <nil>", vm.chunk.Lookups[vm.ip].Location())
			}
			property := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.Push(vm.accessProperty(object, property))
			vm.ip++
		case OpOptionalChain:
			object := vm.stack.Pop()
			if object == nil {
				vm.stack.Push(nil)
				vm.ip = vm.chunk.Instructions[vm.ip+1] - 1
				break
			}
			property := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]].(string)
			vm.stack.Push(vm.accessProperty(object, property))
			vm.ip += 2
		case OpElvis:
			left := vm.stack.Pop()
			if left != nil {
				vm.stack.Push(left)
				jumpAhead := vm.chunk.Instructions[vm.ip+1]
				vm.ip += jumpAhead
			} else {
				vm.ip++
			}
		case OpTernary:
			condition := vm.stack.Pop()
			if CastToBool(condition) {
				vm.ip++
			} else {
				jumpToFalse := vm.chunk.Instructions[vm.ip+1]
				vm.ip += jumpToFalse
			}
		case OpJmp:
			jump := vm.chunk.Instructions[vm.ip+1]
			vm.ip += jump
		case OpArrayAccess:
			_index := vm.stack.Pop()
			_value := vm.stack.Pop()
			value := reflect.ValueOf(_value)
			switch value.Kind() {
			case reflect.Array, reflect.Slice:
				result := castInt(_index)
				if _, ok := result.(error); ok {
					return nil, errors2.New(fmt.Sprintf("[%s] expected to produce an Integer, got %v", vm.chunk.Lookups[vm.ip].(*FieldAccess).Index.Literal(), reflect.TypeOf(_index)), vm.chunk.Lookups[vm.ip].Location())
				}
				vm.stack.Push(value.Index(result.(int)).Interface())
			case reflect.Map:
				vm.stack.Push(value.MapIndex(reflect.ValueOf(_index)).Interface())
			case reflect.Struct:
				index, ok := _index.(string)
				if !ok {
					return nil, errors2.New(fmt.Sprintf("struct field accessor expected to be string, got %v", reflect.TypeOf(_index)), vm.chunk.Lookups[vm.ip].Location())
				}
				vm.stack.Push(value.FieldByName(index).Interface())
			default:
				return nil, errors2.New(fmt.Sprintf("expected array or object, got %v", reflect.TypeOf(_value)), vm.chunk.Lookups[vm.ip].Location())
			}
		case OpBuiltin1:
			function := builtinsOne[vm.chunk.Instructions[vm.ip+1]]
			arg := vm.stack.Pop()
			result := function(arg)

			if general, ok := result.(error); ok {
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
			}

			vm.stack.Push(result)
			vm.ip++
		case OpBuiltin2:
			function := builtinsTwo[vm.chunk.Instructions[vm.ip+1]]
			arg2 := vm.stack.Pop()
			arg1 := vm.stack.Pop()
			result := function(arg1, arg2)
			if general, ok := result.(error); ok {
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
			}
			vm.stack.Push(result)
			vm.ip++
		case OpBuiltin3:
			function := builtinsThree[vm.chunk.Instructions[vm.ip+1]]
			arg3 := vm.stack.Pop()
			arg2 := vm.stack.Pop()
			arg1 := vm.stack.Pop()
			result := function(arg1, arg2, arg3)
			if general, ok := result.(error); ok {
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
			}
			vm.stack.Push(result)
			vm.ip++
		case OpArray:
			count := vm.chunk.Instructions[vm.ip+1]
			items := make([]any, count)
			for j := 0; j < count; j++ {
				items[count-j-1] = vm.stack.Pop()
			}
			vm.stack.Push(items)
			vm.ip++

		case OpCall:
			argumentCount := vm.chunk.Instructions[vm.ip+1]
			vm.ip++
			args := make([]reflect.Value, argumentCount)
			for j := argumentCount - 1; j >= 0; j-- {
				args[j] = reflect.ValueOf(vm.stack.Pop())
			}

			fn := vm.stack.Pop()
			reflectedFunction := reflect.ValueOf(fn)
			if !reflectedFunction.IsValid() || reflectedFunction.Kind() != reflect.Func {
				vm.currentError = errors2.New(fmt.Sprintf("expected function, got %v", reflect.TypeOf(fn)), vm.chunk.Lookups[vm.ip-1].Location())
				break
			}
			results := reflectedFunction.Call(args)
			if len(results) == 1 {
				vm.stack.Push(results[0].Interface())
			} else if len(results) > 1 {
				vm.stack.Push(reflectedSliceToInterfaceSlice(results))
			}
		case OpGet:
			ident := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.Push(env[ident])
			vm.ip++
		case OpConstant:
			constant := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]]
			vm.ip++
			vm.stack.Push(constant)
		case OpIn:
			right := vm.stack.Pop()
			left := vm.stack.Pop()

			for i := 0; i < reflect.ValueOf(right).Len(); i++ {
				if reflect.ValueOf(right).Index(i).Interface() == left {
					vm.stack.Push(true)
					continue outerLoop
				}
			}
			vm.stack.Push(false)
		case OpNil:
			vm.stack.Push(nil)
		case OpEq:
			left := vm.stack.Pop()
			right := vm.stack.Pop()
			vm.stack.Push(left == right)
		case OpNegate:
			vm.stack.Push(negate(vm.stack.Pop()))
		case OpNeq:
			right := vm.stack.Pop()
			left := vm.stack.Pop()
			vm.stack.Push(left != right)
		case OpNot:
			vm.stack.Push(!CastToBool(vm.stack.Pop()))
		case OpAdd:
			vm.executeInfixExpression(binaryAddition)
		case OpLt:
			vm.executeInfixExpression(binaryLessThan)
		case OpGt:
			vm.executeInfixExpression(binaryGreaterThan)
		case OpGte:
			vm.executeInfixExpression(binaryGreaterThanEqual)
		case OpLte:
			vm.executeInfixExpression(binaryLessThanEqual)
		case OpSubtract:
			vm.executeInfixExpression(binarySubtraction)
		case OpMultiply:
			vm.executeInfixExpression(binaryMultiplication)
		case OpDivide:
			vm.executeInfixExpression(binaryDivision)
		case OpModulo:
			vm.executeInfixExpression(binaryModulo)
		case OpPower:
			vm.executeInfixExpression(binaryExponentiation)
		case OpAnd:
			vm.executeInfixExpression(and)
		case OpOr:
			vm.executeInfixExpression(or)
		case OpCodeCount:
			panic("unreachable")
		}
		if vm.currentError != nil {
			return nil, vm.currentError
		}
	}

	if len(vm.stack) == 0 {
		return nil, errors2.New("expression does not return a value", vm.chunk.Lookups[0].Location())
	}

	if len(vm.stack) != 1 {
		return nil, fmt.Errorf("expression returns multiple values")
	}

	return vm.stack.Pop(), nil
}

func (vm *VM) executeInfixExpression(fn func(any, any) any) {
	right := vm.stack.Pop()
	left := vm.stack.Pop()
	res := fn(left, right)
	if general, ok := res.(error); ok {
		vm.currentError = errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].(*InfixExpression).Token.Location)
		return
	}
	vm.stack.Push(res)
}

func (vm *VM) accessProperty(base any, property string) any {
	value := reflect.ValueOf(base)
	if !value.IsValid() {
		return nil
	}

	var reflected reflect.Value
	switch value.Kind() {
	case reflect.Map:
		reflected = value.MapIndex(reflect.ValueOf(property))
	case reflect.Struct:
		reflected = value.FieldByName(property)
		if !reflected.IsValid() {
			reflected = value.MethodByName(property)
		}
	case reflect.Pointer:
		if value.Elem().Kind() == reflect.Struct {
			reflected = value.Elem().FieldByName(property)
			if !reflected.IsValid() {
				reflected = value.MethodByName(property)
			}
			if reflected.IsValid() {
				return reflected.Interface()
			}
		}
		return vm.accessProperty(value.Elem().Interface(), property)
	default:
		reflected = value.MethodByName(property)
		if reflected.IsValid() {
			return reflected.Interface()
		}
		ptrBaseValue := reflect.New(value.Type())
		ptrBaseValue.Elem().Set(value)
		methodValue := ptrBaseValue.MethodByName(property)
		if methodValue.IsValid() {
			return methodValue.Interface()
		}
	}
	if !reflected.IsValid() {
		return nil
	}

	return reflected.Interface()
}

func reflectedSliceToInterfaceSlice(vs []reflect.Value) []interface{} {
	is := make([]interface{}, len(vs))
	for i, v := range vs {
		is[i] = v.Interface()
	}
	return is
}

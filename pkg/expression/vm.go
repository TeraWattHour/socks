package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"reflect"
)

type VM struct {
	chunk        Chunk
	stack        Stack
	funcs        map[string]any
	ip           int
	currentError error
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
		vm.currentError = nil
		switch vm.chunk.Instructions[vm.ip] {
		case OpChain:
			object := vm.stack.pop()
			if object == nil {
				return nil, errors2.New("cannot access properties of <nil>", vm.chunk.Lookups[vm.ip].Location())
			}
			property := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+1]].(string)
			vm.stack.push(vm.accessProperty(object, property))
			vm.ip++
		case OpOptionalChain:
			object := vm.stack.pop()
			if object == nil {
				vm.stack.push(nil)
				vm.ip = vm.chunk.Instructions[vm.ip+1] - 1
				break
			}
			property := vm.chunk.Constants[vm.chunk.Instructions[vm.ip+2]].(string)
			vm.stack.push(vm.accessProperty(object, property))
			vm.ip += 2
		case OpElvis:
			left := vm.stack.pop()
			if left != nil {
				vm.stack.push(left)
				jumpAhead := vm.chunk.Instructions[vm.ip+1]
				vm.ip += jumpAhead
			} else {
				vm.ip++
			}
		case OpTernary:
			condition := vm.stack.pop()
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
			_index := vm.stack.pop()
			_value := vm.stack.pop()
			value := reflect.ValueOf(_value)
			switch value.Kind() {
			case reflect.Array, reflect.Slice:
				result := castInt(_index)
				if _, ok := result.(error); ok {
					return nil, errors2.New(fmt.Sprintf("[%s] expected to produce an Integer, got %v", vm.chunk.Lookups[vm.ip].(*FieldAccess).Index.Literal(), reflect.TypeOf(_index)), vm.chunk.Lookups[vm.ip].Location())
				}
				vm.stack.push(value.Index(result.(int)).Interface())
			case reflect.Map:
				vm.stack.push(value.MapIndex(reflect.ValueOf(_index)).Interface())
			case reflect.Struct:
				index, ok := _index.(string)
				if !ok {
					return nil, errors2.New(fmt.Sprintf("struct field accessor expected to be string, got %v", reflect.TypeOf(_index)), vm.chunk.Lookups[vm.ip].Location())
				}
				vm.stack.push(value.FieldByName(index).Interface())
			default:
				return nil, errors2.New(fmt.Sprintf("expected array or object, got %v", reflect.TypeOf(_value)), vm.chunk.Lookups[vm.ip].Location())
			}
		case OpBuiltin1:
			function := builtinsOne[vm.chunk.Instructions[vm.ip+1]]
			arg := vm.stack.pop()
			result := function(arg)

			if general, ok := result.(error); ok {
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
			}

			vm.stack.push(result)
			vm.ip++
		case OpBuiltin2:
			function := builtinsTwo[vm.chunk.Instructions[vm.ip+1]]
			arg2 := vm.stack.pop()
			arg1 := vm.stack.pop()
			result := function(arg1, arg2)
			if general, ok := result.(error); ok {
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
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
				return nil, errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].Location())
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
			for j := argumentCount - 1; j >= 0; j-- {
				args[j] = reflect.ValueOf(vm.stack.pop())
			}

			fn := vm.stack.pop()
			reflectedFunction := reflect.ValueOf(fn)
			if !reflectedFunction.IsValid() || reflectedFunction.Kind() != reflect.Func {
				vm.currentError = errors2.New(fmt.Sprintf("expected function, got %v", reflect.TypeOf(fn)), vm.chunk.Lookups[vm.ip-1].Location())
				break
			}
			results := reflectedFunction.Call(args)
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
		case OpNil:
			vm.stack.push(nil)
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

	return vm.stack.pop(), nil
}

func (vm *VM) executeInfixExpression(fn func(any, any) any) {
	right := vm.stack.pop()
	left := vm.stack.pop()
	res := fn(left, right)
	if general, ok := res.(error); ok {
		vm.currentError = errors2.New(general.Error(), vm.chunk.Lookups[vm.ip].(*InfixExpression).Token.LocationStart)
		return
	}
	vm.stack.push(res)
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

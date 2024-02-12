package expression

import (
	"fmt"
	"reflect"
)

type VM struct {
	chunk Chunk
	stack Stack
	funcs map[string]any
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
	chain := false
outerLoop:
	for ip := 0; ip < len(vm.chunk.Instructions); ip++ {
		switch vm.chunk.Instructions[ip] {
		case OpChain:
			chain = true
		case OpArrayAccess:
			_index := vm.stack.pop()
			_value := vm.stack.pop()
			val := reflect.ValueOf(_value)
			switch val.Kind() {
			case reflect.Array, reflect.Slice:
				if !reflect.ValueOf(_index).CanConvert(reflect.TypeOf(0)) {
					return nil, fmt.Errorf("expected int, got %v", reflect.TypeOf(_index))
				}
				index := reflect.ValueOf(_index).Convert(reflect.TypeOf(0)).Int()

				vm.stack.push(val.Index(int(index)).Interface())
			case reflect.Map:
				vm.stack.push(val.MapIndex(reflect.ValueOf(_index)).Interface())
			default:
				panic(fmt.Sprintf("expected array or object, got %v", reflect.TypeOf(_value)))
			}
		case OpBuiltin1:
			function := builtinsOne[builtinNames[vm.chunk.Instructions[ip+1]]]
			arg := vm.stack.pop()
			vm.stack.push(function(arg))
			ip++
		case OpBuiltin2:
			function := builtinsTwo[builtinNames[vm.chunk.Instructions[ip+1]]]
			arg2 := vm.stack.pop()
			arg1 := vm.stack.pop()
			vm.stack.push(function(arg1, arg2))
			ip++
		case OpBuiltin3:
			function := builtinsThree[builtinNames[vm.chunk.Instructions[ip+1]]]
			arg3 := vm.stack.pop()
			arg2 := vm.stack.pop()
			arg1 := vm.stack.pop()
			vm.stack.push(function(arg1, arg2, arg3))
			ip++
		case OpArray:
			count := vm.chunk.Instructions[ip+1]
			ip++
			items := make([]any, count)
			for j := 0; j < count; j++ {
				items[count-j-1] = vm.stack.pop()
			}
			vm.stack.push(items)

		case OpCall:
			argumentCount := vm.chunk.Instructions[ip+1]
			ip++
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
			if chain {
				value := vm.stack.pop()
				ident := vm.chunk.Constants[vm.chunk.Instructions[ip+1]]
				vm.stack.push(accessVariable(value, ident))
				chain = false
			} else {
				ident := vm.chunk.Constants[vm.chunk.Instructions[ip+1]].(string)
				vm.stack.push(env[ident])
			}
			ip++
		case OpConstant:
			constant := vm.chunk.Constants[vm.chunk.Instructions[ip+1]]
			ip++
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
		case OpEqual:
			left := vm.stack.pop()
			right := vm.stack.pop()
			vm.stack.push(left == right)
		case OpNotEqual:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(left != right)
		case OpNot:
			last := vm.stack.pop().(bool)
			vm.stack.push(!last)
		case OpAdd:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryAddition(left, right))
		case OpLessThan:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryLessThan(left, right))
		case OpGreaterThan:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryGreaterThan(left, right))
		case OpGreaterThanOrEqual:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryGreaterThanEqual(left, right))
		case OpLessThanOrEqual:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryLessThanEqual(left, right))
		case OpSubtract:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binarySubtraction(left, right))
		case OpMultiply:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryMultiplication(left, right))
		case OpDivide:
			right := vm.stack.pop()
			left := vm.stack.pop()
			vm.stack.push(binaryDivision(left, right))
		}
	}

	if len(vm.stack) == 0 {
		return nil, fmt.Errorf("expression doesnt return a value")
	}

	if len(vm.stack) != 1 {
		fmt.Println(vm.stack)
		return nil, fmt.Errorf("expression returns multiple values")
	}

	return vm.stack.pop(), nil
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

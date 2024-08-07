package expression

import (
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"reflect"
)

type VM struct {
	program      Program
	stack        helpers.Stack[any]
	ip           int
	currentError error
}

func NewVM(program Program) *VM {
	return &VM{
		program: program,
		stack:   []any{},
	}
}

func (vm *VM) Run(env map[string]any) (any, error) {
	if vm == nil {
		return nil, nil
	}

outerLoop:
	for vm.ip = 0; vm.ip < len(vm.program.Instructions); vm.ip++ {
		vm.currentError = nil
		switch vm.program.Instructions[vm.ip] {
		case OpChain:
			object := vm.stack.Pop()
			if object == nil {
				return nil, vm.error("can't access properties of <nil>", vm.program.Lookups[vm.ip].Location())
			}
			property := vm.program.Constants[vm.takeNext()].(string)
			vm.stack.Push(vm.accessProperty(object, property))
		case OpOptionalChain:
			object := vm.stack.Pop()
			if object == nil {
				vm.stack.Push(nil)
				vm.ip += vm.nextInstruction()
			} else {
				vm.stack.Push(object)
				vm.ip++
			}
		case OpElvis:
			left := vm.stack.Pop()
			if left != nil {
				vm.stack.Push(left)
				vm.ip += vm.nextInstruction()
			} else {
				vm.ip++
			}
		case OpTernary:
			condition := vm.stack.Pop()
			if CastToBool(condition) {
				vm.ip++
			} else {
				vm.ip += vm.nextInstruction()
			}
		case OpPop:
			vm.stack.Pop()
		case OpJmp:
			vm.ip += vm.nextInstruction()
		case OpPropertyAccess:
			_index := vm.stack.Pop()
			_value := vm.stack.Pop()
			value := reflect.ValueOf(_value)
			lookup := vm.program.Lookups[vm.ip].(*FieldAccess)
			switch value.Kind() {
			case reflect.Array, reflect.Slice:
				result := castInt(_index)
				if err, ok := result.(error); ok {
					return nil, vm.error("forbidden array index access, "+err.Error(), lookup.Index.Location())
				}
				vm.stack.Push(value.Index(result.(int)).Interface())
			case reflect.Map:
				vm.stack.Push(value.MapIndex(reflect.ValueOf(_index)).Interface())
			case reflect.Struct:
				index, ok := _index.(string)
				if !ok {
					return nil, vm.error(fmt.Sprintf("struct field accessor must be of type string, got %T", _index), lookup.Index.Location())
				}
				vm.stack.Push(value.FieldByName(index).Interface())
			default:
				return nil, vm.error(fmt.Sprintf("forbidden access of properties of %T", _value), lookup.Location())
			}
		case OpArray:
			count := vm.program.Instructions[vm.ip+1]
			items := make([]any, count)
			for j := 0; j < count; j++ {
				items[count-j-1] = vm.stack.Pop()
			}
			vm.stack.Push(items)
			vm.ip++

		case OpCall:
			argumentCount := vm.takeNext()

			args := make([]reflect.Value, argumentCount)
			for j := argumentCount - 1; j >= 0; j-- {
				args[j] = reflect.ValueOf(vm.stack.Pop())
			}

			fn := vm.stack.Pop()
			reflectedFunction := reflect.ValueOf(fn)
			if !reflectedFunction.IsValid() || reflectedFunction.Kind() != reflect.Func {
				vm.currentError = vm.error(fmt.Sprintf("can't call %T", fn), vm.program.Lookups[vm.ip-1].(*FunctionCall).Location())
				break
			}
			results := reflectedFunction.Call(args)
			if len(results) == 1 {
				result := results[0].Interface()
				switch result := result.(type) {
				case *castError:
					vm.currentError = vm.error(result.Error(), vm.program.Lookups[vm.ip-1].(*FunctionCall).Location())
				default:
					vm.stack.Push(result)
				}
			} else if len(results) > 1 {
				vm.stack.Push(reflectedSliceToInterfaceSlice(results))
			}
		case OpGet:
			ident := vm.program.Constants[vm.takeNext()].(string)
			if env[ident] != nil {
				vm.stack.Push(env[ident])
			} else if f, ok := builtinsOne[ident]; ok {
				vm.stack.Push(f)
			} else {
				vm.stack.Push(nil)
			}
		case OpConstant:
			vm.stack.Push(vm.program.Constants[vm.takeNext()])
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
			vm.executeInfixExpression(operationAddition)
		case OpLt:
			vm.executeInfixExpression(operationLess)
		case OpGt:
			vm.executeInfixExpression(operationGreater)
		case OpGte:
			vm.executeInfixExpression(operationGreaterEqual)
		case OpLte:
			vm.executeInfixExpression(operationLessEqual)
		case OpSubtract:
			vm.executeInfixExpression(operationSubtraction)
		case OpMultiply:
			vm.executeInfixExpression(operationMultiplication)
		case OpDivide:
			vm.executeInfixExpression(operationDivision)
		case OpModulo:
			vm.executeInfixExpression(operationModulus)
		case OpPower:
			vm.executeInfixExpression(operationExponentiation)
		case OpAnd:
			vm.executeInfixExpression(and)
		case OpOr:
			vm.executeInfixExpression(or)
		default:
			panic("unreachable")
		}
		if vm.currentError != nil {
			return nil, vm.currentError
		}

	}

	if len(vm.stack) == 0 {
		return nil, vm.error("expression does not return a value", vm.program.Lookups[0].Location())
	}

	if len(vm.stack) != 1 {
		return nil, vm.error("expression returns multiple values", vm.program.Lookups[0].Location())
	}

	return vm.stack.Pop(), nil
}

func (vm *VM) executeInfixExpression(fn func(any, any) any) {
	right := vm.stack.Pop()
	left := vm.stack.Pop()
	res := fn(left, right)
	if general, ok := res.(error); ok {
		vm.currentError = vm.error(general.Error(), vm.program.Lookups[vm.ip].Location())
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

func (vm *VM) takeNext() int {
	vm.ip++
	return vm.program.Instructions[vm.ip]
}

func (vm *VM) nextInstruction() int {
	return vm.program.Instructions[vm.ip+1]
}

func (vm *VM) error(message string, location helpers.Location) error {
	return errors2.New(message, location)
}

func reflectedSliceToInterfaceSlice(vs []reflect.Value) []interface{} {
	is := make([]interface{}, len(vs))
	for i, v := range vs {
		is[i] = v.Interface()
	}
	return is
}

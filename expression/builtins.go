package expression

import (
	"fmt"
	"reflect"
	"slices"
)

// Order of builtins is important
var builtinNames = []string{
	// One-argument builtins
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
	// Two-argument builtins

	// Three-argument builtins
	"range",
}

var builtinTypes = map[string][]string{
	"float32": {"Numeric", "float32"},
	"float64": {"Numeric", "float64"},
	"int":     {"Numeric", "int"},
	"int8":    {"Numeric", "int8"},
	"int16":   {"Numeric", "int16"},
	"int32":   {"Numeric", "int32"},
	"int64":   {"Numeric", "int64"},
	"uint":    {"Numeric", "uint"},
	"uint8":   {"Numeric", "uint8"},
	"uint16":  {"Numeric", "uint16"},
	"uint32":  {"Numeric", "uint32"},
	"uint64":  {"Numeric", "uint64"},
	"uintptr": {"Numeric", "uintptr"},
	"len":     {"Countable", "len"},
	"range":   {"Integer", "Integer", "Integer = 1", "[]int"},
}

var builtinsOne = []func(any) any{
	castFloat32,
	castFloat64,
	castInt,
	castInt8,
	castInt16,
	castInt32,
	castInt64,
	castUint,
	castUint8,
	castUint16,
	castUint32,
	castUint64,
	castUintptr,
	length,
}

var builtinsTwo = []func(any, any) any{}

var builtinsThree = []func(any, any, any) any{
	rangeArray,
}

var numBuiltinsOne = reflect.ValueOf(builtinsOne).Len()
var numBuiltinsTwo = reflect.ValueOf(builtinsTwo).Len()
var numBuiltinsThree = reflect.ValueOf(builtinsThree).Len()

func builtinRelativeIndex(name string) int {
	idx := slices.Index(builtinNames, name)
	if idx == -1 {
		return -1
	}

	if idx < numBuiltinsOne {
		return idx
	} else if idx < numBuiltinsOne+numBuiltinsTwo {
		return idx - numBuiltinsOne
	} else if idx < numBuiltinsOne+numBuiltinsTwo+numBuiltinsThree {
		return idx - numBuiltinsOne - numBuiltinsTwo
	}

	return -1
}

func builtinType(name string) int {
	idx := slices.Index(builtinNames, name)
	if idx == -1 {
		return -1
	}

	if idx < numBuiltinsOne {
		return 1
	} else if idx < numBuiltinsOne+numBuiltinsTwo {
		return 2
	} else if idx < numBuiltinsOne+numBuiltinsTwo+numBuiltinsThree {
		return 3
	}

	return -1
}

func length(_val any) any {
	switch val := _val.(type) {
	case string:
		return len(val)
	case []any:
		return len(val)
	case map[any]any:
		return len(val)
	}
	return reflect.ValueOf(_val).Len()
}

func rangeArray(_start, _end, _step any) any {
	start, startOk := castInt(_start).(int)
	end, endOk := castInt(_end).(int)
	step, stepOk := castInt(_step).(int)
	if !startOk || !endOk || !stepOk {
		return fmt.Errorf("call to range(%T, %T, %T) -> []int does not match the signature of rangeStep(Integer, Integer, Integer = 1) -> []int", _start, _end, _step)
	}

	if step == 0 {
		return fmt.Errorf("step cannot be 0")
	}
	if start < end && step < 0 {
		return fmt.Errorf("step cannot be negative while start < end")
	}
	if start > end && step > 0 {
		return fmt.Errorf("step cannot be positive while start > end")
	}
	var result []int
	for i := start; i < end; i += step {
		result = append(result, i)
	}
	return result
}

func negate(val any) any {
	switch val := val.(type) {
	case int:
		return -val
	case int8:
		return -val
	case int16:
		return -val
	case int32:
		return -val
	case int64:
		return -val
	case float32:
		return -val
	case float64:
		return -val
	case string:
		runes := []rune(val)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}
	return fmt.Errorf("cannot negate %T", val)
}

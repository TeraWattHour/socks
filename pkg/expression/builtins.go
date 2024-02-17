package expression

import (
	"fmt"
	"reflect"
	"slices"
)

// WARNING: Order of builtins is important
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
	"range":   {"Integer", "Integer", "Integer", "[]int"},
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
	start, startOk := _start.(int)
	end, endOk := _end.(int)
	step, stepOk := _step.(int)
	if !startOk || !endOk || !stepOk {
		return fmt.Errorf("call to rangeStep(%s, %s, %s) -> []int does not match the signature of rangeStep(Integer, Integer, Integer = 1) -> []int", reflect.TypeOf(_start), reflect.TypeOf(_end), reflect.TypeOf(_step))
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
	return fmt.Errorf("cannot negate %s", reflect.TypeOf(val))
}

// BEGIN CASTS
func castInt(val any) any {
	switch val := val.(type) {
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case uint:
		return int(val)
	case uint8:
		return int(val)
	case uint16:
		return int(val)
	case uint32:
		return int(val)
	case uint64:
		return int(val)
	case uintptr:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	}

	return fmt.Errorf("cannot cast %s to int", reflect.TypeOf(val))
}

func castInt8(val any) any {
	switch val := val.(type) {
	case int:
		return int8(val)
	case int8:
		return val
	case int16:
		return int8(val)
	case int32:
		return int8(val)
	case int64:
		return int8(val)
	case uint:
		return int8(val)
	case uint8:
		return int8(val)
	case uint16:
		return int8(val)
	case uint32:
		return int8(val)
	case uint64:
		return int8(val)
	case uintptr:
		return int8(val)
	case float32:
		return int8(val)
	case float64:
		return int8(val)
	}
	return fmt.Errorf("cannot cast %s to int8", reflect.TypeOf(val))
}

func castInt16(val any) any {
	switch val := val.(type) {
	case int:
		return int16(val)
	case int8:
		return int16(val)
	case int16:
		return val
	case int32:
		return int16(val)
	case int64:
		return int16(val)
	case uint:
		return int16(val)
	case uint8:
		return int16(val)
	case uint16:
		return int16(val)
	case uint32:
		return int16(val)
	case uint64:
		return int16(val)
	case uintptr:
		return int16(val)
	case float32:
		return int16(val)
	case float64:
		return int16(val)
	}
	return fmt.Errorf("cannot cast %s to int16", reflect.TypeOf(val))
}

func castInt32(val any) any {
	switch val := val.(type) {
	case int:
		return int32(val)
	case int8:
		return int32(val)
	case int16:
		return int32(val)
	case int32:
		return val
	case int64:
		return int32(val)
	case uint:
		return int32(val)
	case uint8:
		return int32(val)
	case uint16:
		return int32(val)
	case uint32:
		return int32(val)
	case uint64:
		return int32(val)
	case uintptr:
		return int32(val)
	case float32:
		return int32(val)
	case float64:
		return int32(val)
	}
	return fmt.Errorf("cannot cast %s to int32", reflect.TypeOf(val))
}

func castInt64(val any) any {
	switch val := val.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint:
		return int64(val)
	case uint8:
		return int64(val)
	case uint16:
		return int64(val)
	case uint32:
		return int64(val)
	case uint64:
		return int64(val)
	case uintptr:
		return int64(val)
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	}
	return fmt.Errorf("cannot cast %s to int64", reflect.TypeOf(val))
}

func castUint(val any) any {
	switch val := val.(type) {
	case int:
		return uint(val)
	case int8:
		return uint(val)
	case int16:
		return uint(val)
	case int32:
		return uint(val)
	case int64:
		return uint(val)
	case uint:
		return val
	case uint8:
		return uint(val)
	case uint16:
		return uint(val)
	case uint32:
		return uint(val)
	case uint64:
		return uint(val)
	case uintptr:
		return uint(val)
	case float32:
		return uint(val)
	case float64:
		return uint(val)
	}
	return fmt.Errorf("cannot cast %s to uint", reflect.TypeOf(val))
}

func castUint8(val any) any {
	switch val := val.(type) {
	case int:
		return uint8(val)
	case int8:
		return uint8(val)
	case int16:
		return uint8(val)
	case int32:
		return uint8(val)
	case int64:
		return uint8(val)
	case uint:
		return uint8(val)
	case uint8:
		return val
	case uint16:
		return uint8(val)
	case uint32:
		return uint8(val)
	case uint64:
		return uint8(val)
	case uintptr:
		return uint8(val)
	case float32:
		return uint8(val)
	case float64:
		return uint8(val)
	}
	return fmt.Errorf("cannot cast %s to uint8", reflect.TypeOf(val))
}

func castUint16(val any) any {
	switch val := val.(type) {
	case int:
		return uint16(val)
	case int8:
		return uint16(val)
	case int16:
		return uint16(val)
	case int32:
		return uint16(val)
	case int64:
		return uint16(val)
	case uint:
		return uint16(val)
	case uint8:
		return uint16(val)
	case uint16:
		return val
	case uint32:
		return uint16(val)
	case uint64:
		return uint16(val)
	case uintptr:
		return uint16(val)
	case float32:
		return uint16(val)
	case float64:
		return uint16(val)
	}
	return fmt.Errorf("cannot cast %s to uint16", reflect.TypeOf(val))
}

func castUint32(val any) any {
	switch val := val.(type) {
	case int:
		return uint32(val)
	case int8:
		return uint32(val)
	case int16:
		return uint32(val)
	case int32:
		return uint32(val)
	case int64:
		return uint32(val)
	case uint:
		return uint32(val)
	case uint8:
		return uint32(val)
	case uint16:
		return uint32(val)
	case uint32:
		return val
	case uint64:
		return uint32(val)
	case uintptr:
		return uint32(val)
	case float32:
		return uint32(val)
	case float64:
		return uint32(val)
	}
	return fmt.Errorf("cannot cast %s to uint32", reflect.TypeOf(val))
}

func castUint64(val any) any {
	switch val := val.(type) {
	case int:
		return uint64(val)
	case int8:
		return uint64(val)
	case int16:
		return uint64(val)
	case int32:
		return uint64(val)
	case int64:
		return uint64(val)
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	case uintptr:
		return uint64(val)
	case float32:
		return uint64(val)
	case float64:
		return uint64(val)
	}
	return fmt.Errorf("cannot cast %s to uint64", reflect.TypeOf(val))
}

func castUintptr(val any) any {
	switch val := val.(type) {
	case int:
		return uintptr(val)
	case int8:
		return uintptr(val)
	case int16:
		return uintptr(val)
	case int32:
		return uintptr(val)
	case int64:
		return uintptr(val)
	case uint:
		return uintptr(val)
	case uint8:
		return uintptr(val)
	case uint16:
		return uintptr(val)
	case uint32:
		return uintptr(val)
	case uint64:
		return uintptr(val)
	case uintptr:
		return val
	case float32:
		return uintptr(val)
	case float64:
		return uintptr(val)
	}
	return fmt.Errorf("cannot cast %s to uintptr", reflect.TypeOf(val))
}

func castFloat32(val any) any {
	switch val := val.(type) {
	case int:
		return float32(val)
	case int8:
		return float32(val)
	case int16:
		return float32(val)
	case int32:
		return float32(val)
	case int64:
		return float32(val)
	case uint:
		return float32(val)
	case uint8:
		return float32(val)
	case uint16:
		return float32(val)
	case uint32:
		return float32(val)
	case uint64:
		return float32(val)
	case uintptr:
		return float32(val)
	case float32:
		return val
	case float64:
		return float32(val)
	}
	return fmt.Errorf("cannot cast %s to float32", reflect.TypeOf(val))
}

func castFloat64(val any) any {
	switch val := val.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case uintptr:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	}
	return fmt.Errorf("cannot cast %s to float64", reflect.TypeOf(val))
}

// END CASTS

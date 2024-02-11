package expression

import (
	"fmt"
	"reflect"
	"slices"
)

var builtinNames = []string{
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
	"range",
	"rangeStep",
}

var builtinsOne = map[string]func(val any) any{
	"float32": castFloat32,
	"float64": castFloat64,
	"int":     castInt,
	"int8":    castInt8,
	"int16":   castInt16,
	"int32":   castInt32,
	"int64":   castInt64,
	"uint":    castUint,
	"uint8":   castUint8,
	"uint16":  castUint16,
	"uint32":  castUint32,
	"uint64":  castUint64,
	"uintptr": castUintptr,
	"len":     length,
}

var builtinsTwo = map[string]func(val1, val2 any) any{
	"range": rangeArray,
}

var builtinsThree = map[string]func(val1, val2, val3 any) any{
	"rangeStep": rangeArrayStep,
}

var numBuiltinsOne = reflect.ValueOf(builtinsOne).Len()
var numBuiltinsTwo = reflect.ValueOf(builtinsTwo).Len()
var numBuiltinsThree = reflect.ValueOf(builtinsThree).Len()

func builtinType(name string) int {
	idx := slices.Index(builtinNames, name)
	if idx == -1 {
		panic("not a builtin")
	} else if idx < numBuiltinsOne {
		return 1
	} else if idx < numBuiltinsOne+numBuiltinsTwo {
		return 2
	} else if idx < numBuiltinsOne+numBuiltinsTwo+numBuiltinsThree {
		return 3
	}

	panic("not implemented")
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

func rangeArray(start, end any) any {
	switch start := start.(type) {
	case int:
		switch end := end.(type) {
		case int:
			var result []int
			for i := start; i < end; i++ {
				result = append(result, i)
			}
			return result
		}
	}
	panic(fmt.Sprintf("cannot range %s to %s", reflect.TypeOf(start), reflect.TypeOf(end)))
}

func rangeArrayStep(_start, _end, _step any) any {
	start, startOk := _start.(int)
	end, endOk := _end.(int)
	step, stepOk := _step.(int)
	if !startOk || !endOk || !stepOk {
		panic(fmt.Sprintf("cannot range %s to %s over %s", reflect.TypeOf(_start), reflect.TypeOf(_end), reflect.TypeOf(_step)))
	}

	if step == 0 {
		panic("step cannot be 0")
	}
	if start < end && step < 0 {
		panic("step cannot be negative")
	}
	if start > end && step > 0 {
		panic("step cannot be positive")
	}
	var result []int
	for i := start; i < end; i += step {
		result = append(result, i)
	}
	return result
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
	panic(fmt.Sprintf("cannot cast %s to int", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to int8", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to int16", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to int32", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to int64", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uint", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uint8", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uint16", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uint32", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uint64", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to uintptr", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to float32", reflect.TypeOf(val)))
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
	panic(fmt.Sprintf("cannot cast %s to float64", reflect.TypeOf(val)))
}

// END CASTS

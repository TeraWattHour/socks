package expression

import (
	"fmt"
	"reflect"
)

var builtinsOne = map[string]any{
	"float32": castFloat32,
	"float64": castFloat64,
	"fnt":     castInt,
	"fnt8":    castInt8,
	"fnt16":   castInt16,
	"fnt32":   castInt32,
	"fnt64":   castInt64,
	"uint":    castUint,
	"uint8":   castUint8,
	"uint16":  castUint16,
	"uint32":  castUint32,
	"uint64":  castUint64,
	"uintptr": castUintptr,
	"length":  length,
	"range":   _range,
}

var builtinNames = reflect.ValueOf(builtinsOne).MapKeys()

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

func _range(_start, _end, _step any) any {
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
	}
	return fmt.Errorf("can't negate %T", val)
}

package expression

import (
	"fmt"
	"math"
	"reflect"
)

// BEGIN BINARY
func binaryAddition(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a + b
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castInt(a).(int) + castInt(b).(int)
		case float32, float64:
			return castFloat64(a).(float64) + castFloat64(b).(float64)
		}
	case float32, float64:
		switch b := b.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castFloat64(a).(float64) + castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v + %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binarySubtraction(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castInt(a).(int) - castInt(b).(int)
		case float32, float64:
			return castFloat64(a).(float64) - castFloat64(b).(float64)
		}
	case float32, float64:
		switch b := b.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castFloat64(a).(float64) - castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v - %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryMultiplication(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castInt(a).(int) * castInt(b).(int)
		case float32, float64:
			return castFloat64(a).(float64) * castFloat64(b).(float64)
		}
	case float32, float64:
		switch b := b.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castFloat64(a).(float64) * castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v * %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryDivision(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castInt(a).(int) / castInt(b).(int)
		case float32, float64:
			return castFloat64(a).(float64) / castFloat64(b).(float64)
		}
	case float32, float64:
		switch b := b.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castFloat64(a).(float64) / castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v / %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryModulo(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return castInt(a).(int) % castInt(b).(int)
		}
	}
	return fmt.Errorf("invalid operation: %v %% %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}

func binaryExponentiation(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return int(math.Pow(castFloat64(a).(float64), castFloat64(b).(float64)))
		case float32, float64:
			return math.Pow(castFloat64(a).(float64), castFloat64(b).(float64))
		}
	case float32, float64:
		switch b := b.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			return math.Pow(castFloat64(a).(float64), castFloat64(b).(float64))
		}
	}
	return fmt.Errorf("invalid operation: %v ** %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}

// END BINARY

// BEGIN EQUALITY
func binaryLessThan(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return castFloat64(a).(float64) < castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v < %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryLessThanEqual(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return castFloat64(a).(float64) <= castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v <= %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryGreaterThan(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return castFloat64(a).(float64) > castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v > %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}
func binaryGreaterThanEqual(a, b any) any {
	switch a := a.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch b := b.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return castFloat64(a).(float64) >= castFloat64(b).(float64)
		}
	}
	return fmt.Errorf("invalid operation: %v >= %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b))
}

// END EQUALITY

func CastToBool(a any) bool {
	if a == nil {
		return false
	}

	switch a := a.(type) {
	case int:
		return a != 0
	case int8:
		return a != 0
	case int16:
		return a != 0
	case int32:
		return a != 0
	case int64:
		return a != 0
	case uint:
		return a != 0
	case uint8:
		return a != 0
	case uint16:
		return a != 0
	case uint32:
		return a != 0
	case uint64:
		return a != 0
	case uintptr:
		return a != 0
	case float32:
		return a != 0
	case float64:
		return a != 0
	case string:
		return a != ""
	case bool:
		return a
	default:
		kind := reflect.TypeOf(a).Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return reflect.ValueOf(a).Len() > 0
		}
	}
	return true
}

func and(a, b any) any {
	switch a := a.(type) {
	case bool:
		switch b := b.(type) {
		case bool:
			return a && b
		}
	}
	aBool := CastToBool(a)
	bBool := CastToBool(b)
	return aBool && bBool
}

func or(a, b any) any {
	switch a := a.(type) {
	case bool:
		switch b := b.(type) {
		case bool:
			return a || b
		}
	}
	aBool := CastToBool(a)
	bBool := CastToBool(b)
	return aBool || bBool
}

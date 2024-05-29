package expression

import (
	"fmt"
	"math"
	"reflect"
)

func operationAddition(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a + b
		}
	case int:
		switch b := b.(type) {
		case int:
			return a + b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a + b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a + b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a + b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a + b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a + b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a + b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a + b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a + b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a + b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a + b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a + b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a + b
		}
	}
	return fmt.Errorf("invalid operation: %v + %v (mismatched types %T and %T)", a, b, a, b)
}

func operationSubtraction(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a - b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a - b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a - b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a - b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a - b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a - b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a - b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a - b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a - b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a - b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a - b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a - b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a - b
		}
	}
	return fmt.Errorf("invalid operation: %v - %v (mismatched types %T and %T)", a, b, a, b)
}

func operationMultiplication(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a * b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a * b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a * b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a * b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a * b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a * b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a * b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a * b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a * b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a * b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a * b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a * b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a * b
		}
	}
	return fmt.Errorf("invalid operation: %v * %v (mismatched types %T and %T)", a, b, a, b)
}

func operationExponentiation(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return int(math.Pow(float64(a), float64(b)))
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return int8(math.Pow(float64(a), float64(b)))
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return int16(math.Pow(float64(a), float64(b)))
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return int32(math.Pow(float64(a), float64(b)))
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return int64(math.Pow(float64(a), float64(b)))
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return uint(math.Pow(float64(a), float64(b)))
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return uint8(math.Pow(float64(a), float64(b)))
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return uint16(math.Pow(float64(a), float64(b)))
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return uint32(math.Pow(float64(a), float64(b)))
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return uint64(math.Pow(float64(a), float64(b)))
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return uintptr(math.Pow(float64(a), float64(b)))
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return float32(math.Pow(float64(a), float64(b)))
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return math.Pow(a, b)
		}
	}
	return fmt.Errorf("invalid operation: %v ** %v (mismatched types %T and %T)", a, b, a, b)
}

func operationDivision(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a / b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a / b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a / b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a / b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a / b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a / b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a / b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a / b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a / b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a / b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a / b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a / b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a / b
		}
	}
	return fmt.Errorf("invalid operation: %v / %v (mismatched types %T and %T)", a, b, a, b)
}

func operationModulus(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a % b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a % b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a % b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a % b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a % b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a % b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a % b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a % b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a % b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a % b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a % b
		}
	}
	return fmt.Errorf("invalid operation: %v %% %v (mismatched types %T and %T)", a, b, a, b)
}

func operationEqual(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a == b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a == b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a == b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a == b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a == b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a == b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a == b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a == b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a == b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a == b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a == b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a == b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a == b
		}
	}
	return fmt.Errorf("invalid operation: %v == %v (mismatched types %T and %T)", a, b, a, b)
}

func operationNotEqual(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a != b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a != b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a != b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a != b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a != b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a != b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a != b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a != b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a != b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a != b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a != b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a != b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a != b
		}
	}
	return fmt.Errorf("invalid operation: %v != %v (mismatched types %T and %T)", a, b, a, b)
}

func operationLess(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a < b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a < b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a < b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a < b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a < b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a < b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a < b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a < b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a < b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a < b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a < b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a < b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a < b
		}
	}
	return fmt.Errorf("invalid operation: %v < %v (mismatched types %T and %T)", a, b, a, b)
}

func operationLessEqual(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a <= b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a <= b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a <= b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a <= b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a <= b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a <= b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a <= b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a <= b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a <= b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a <= b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a <= b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a <= b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a <= b
		}
	}
	return fmt.Errorf("invalid operation: %v <= %v (mismatched types %T and %T)", a, b, a, b)
}

func operationGreater(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a > b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a > b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a > b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a > b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a > b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a > b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a > b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a > b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a > b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a > b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a > b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a > b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a > b
		}
	}
	return fmt.Errorf("invalid operation: %v > %v (mismatched types %T and %T)", a, b, a, b)
}

func operationGreaterEqual(a, b any) any {
	switch a := a.(type) {
	case int:
		switch b := b.(type) {
		case int:
			return a >= b
		}
	case int8:
		switch b := b.(type) {
		case int8:
			return a >= b
		}
	case int16:
		switch b := b.(type) {
		case int16:
			return a >= b
		}
	case int32:
		switch b := b.(type) {
		case int32:
			return a >= b
		}
	case int64:
		switch b := b.(type) {
		case int64:
			return a >= b
		}
	case uint:
		switch b := b.(type) {
		case uint:
			return a >= b
		}
	case uint8:
		switch b := b.(type) {
		case uint8:
			return a >= b
		}
	case uint16:
		switch b := b.(type) {
		case uint16:
			return a >= b
		}
	case uint32:
		switch b := b.(type) {
		case uint32:
			return a >= b
		}
	case uint64:
		switch b := b.(type) {
		case uint64:
			return a >= b
		}
	case uintptr:
		switch b := b.(type) {
		case uintptr:
			return a >= b
		}
	case float32:
		switch b := b.(type) {
		case float32:
			return a >= b
		}
	case float64:
		switch b := b.(type) {
		case float64:
			return a >= b
		}
	}
	return fmt.Errorf("invalid operation: %v >= %v (mismatched types %T and %T)", a, b, a, b)
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

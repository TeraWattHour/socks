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
	panic(fmt.Sprintf("invalid operation: %v + %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}
func binarySubtraction(a, b any) any {
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
	panic(fmt.Sprintf("invalid operation: %v - %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}
func binaryMultiplication(a, b any) any {
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
	panic(fmt.Sprintf("invalid operation: %v * %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}
func binaryDivision(a, b any) any {
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
	panic(fmt.Sprintf("invalid operation: %v / %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}
func binaryModulo(a, b any) any {
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
	panic(fmt.Sprintf("invalid operation: %v %% %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}

func binaryExponentiation(a, b any) any {
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
	panic(fmt.Sprintf("invalid operation: %v == %v (mismatched types %s and %s)", a, b, reflect.TypeOf(a), reflect.TypeOf(b)))
}

// END BINARY

// BEGIN EQUALITY
func binaryEqual(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a == b
		}
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
	return reflect.DeepEqual(a, b)
}
func binaryNotEqual(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a != b
		}
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
	return reflect.DeepEqual(a, b)
}
func binaryLessThan(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a < b
		}
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
	return reflect.DeepEqual(a, b)
}
func binaryLessThanEqual(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a <= b
		}
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
	return reflect.DeepEqual(a, b)
}
func binaryGreaterThan(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a > b
		}
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
	return reflect.DeepEqual(a, b)
}
func binaryGreaterThanEqual(a, b any) any {
	switch a := a.(type) {
	case string:
		switch b := b.(type) {
		case string:
			return a >= b
		}
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
	return reflect.DeepEqual(a, b)
}

// END EQUALITY

package expression

import "fmt"

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

	return fmt.Errorf("cannot cast %T to int", val)
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
	return fmt.Errorf("cannot cast %T to int8", val)
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
	return fmt.Errorf("cannot cast %T to int16", val)
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
	return fmt.Errorf("cannot cast %T to int32", val)
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
	return fmt.Errorf("cannot cast %T to int64", val)
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
	return fmt.Errorf("cannot cast %T to uint", val)
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
	return fmt.Errorf("cannot cast %T to uint8", val)
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
	return fmt.Errorf("cannot cast %T to uint16", val)
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
	return fmt.Errorf("cannot cast %T to uint32", val)
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
	return fmt.Errorf("cannot cast %T to uint64", val)
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
	return fmt.Errorf("cannot cast %T to uintptr", val)
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
	return fmt.Errorf("cannot cast %T to float32", val)
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
	return fmt.Errorf("cannot cast %T to float64", val)
}

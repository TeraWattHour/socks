package expression

import "fmt"

type castError struct {
	from string
	to   string
}

func cerr(from any, to string) *castError {
	return &castError{from: fmt.Sprintf("%T", from), to: to}
}

func (e *castError) Error() string {
	return fmt.Sprintf("can't cast %s to %s", e.from, e.to)
}

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

	return cerr(val, "int")
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
	return cerr(val, "int8")
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
	return cerr(val, "int16")
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
	return cerr(val, "int32")
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
	return cerr(val, "int64")
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
	return cerr(val, "uint")
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
	return cerr(val, "uint8")
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
	return cerr(val, "uint16")
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
	return cerr(val, "uint32")
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
	return cerr(val, "uint64")
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
	return cerr(val, "uintptr")
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
	return cerr(val, "float32")
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
	return cerr(val, "float64")
}

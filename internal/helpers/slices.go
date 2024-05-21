package helpers

import (
	"reflect"
)

func ConvertInterfaceToSlice(result chan any, obj any) {
	sliceValue := reflect.ValueOf(obj)

	switch sliceValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < sliceValue.Len(); i++ {
			result <- sliceValue.Index(i).Interface()
		}
	case reflect.Map:
		for _, key := range sliceValue.MapKeys() {
			result <- sliceValue.MapIndex(key).Interface()
		}
	default:
		panic("unreachable")
	}

}

func IsIterable(obj any) bool {
	value := reflect.ValueOf(obj)
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array || value.Kind() == reflect.Map
}

// Subset checks whether a is a subset of B
func Subset[T comparable](a, B []T) bool {
	if len(a) > len(B) {
		return false
	}

	for _, v := range a {
		found := false
		for _, w := range B {
			if v == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type Stack[T any] []T

func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[T]) Pop() T {
	if len(*s) == 0 {
		var noop T
		return noop
	}
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

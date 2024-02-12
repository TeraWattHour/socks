package helpers

import (
	"reflect"
)

func ConvertInterfaceToSlice(obj any) []any {
	sliceValue := reflect.ValueOf(obj)

	if sliceValue.Kind() == reflect.Slice || sliceValue.Kind() == reflect.Array {
		resultSlice := make([]any, sliceValue.Len())
		for i := 0; i < sliceValue.Len(); i++ {
			resultSlice[i] = sliceValue.Index(i).Interface()
		}

		return resultSlice
	} else if sliceValue.Kind() == reflect.Map {
		resultSlice := make([]any, sliceValue.Len())
		for i, key := range sliceValue.MapKeys() {
			resultSlice[i] = sliceValue.MapIndex(key).Interface()
		}

		return resultSlice
	}

	return nil
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		found := true
		for _, w := range b {
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

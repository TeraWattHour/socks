package helpers

import (
	"reflect"
)

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Map[T any, R any](previous []T, fn func(T) R) []R {
	result := make([]R, len(previous))

	for i, item := range previous {
		result[i] = fn(item)
	}

	return result
}

func ConvertInterfaceToSlice(obj interface{}) []interface{} {
	sliceValue := reflect.ValueOf(obj)

	if sliceValue.Kind() != reflect.Slice {
		return nil
	}

	resultSlice := make([]interface{}, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		value := reflect.ValueOf(sliceValue.Index(i)).Interface()
		resultSlice[i] = value
	}

	return resultSlice
}

package helpers

import (
	"reflect"
)

func CombineSlicesUnique[T comparable](slices ...[]T) []T {
	result := make([]T, 0)

	for _, slice := range slices {
		for _, item := range slice {
			if !Contains(result, item) {
				result = append(result, item)
			}
		}
	}

	return result
}

func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0)

	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}

	return result
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
		resultSlice[i] = sliceValue.Index(i).Interface()
	}

	return resultSlice
}

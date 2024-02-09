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

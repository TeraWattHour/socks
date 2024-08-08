package helpers

import (
	"reflect"
	"slices"
)

type Key any
type Value any

type KeyValuePair struct {
	Key
	Value
}

type tuple = KeyValuePair

func ExtractValues(result chan tuple, obj any) {
	sliceValue := reflect.ValueOf(obj)

	switch sliceValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < sliceValue.Len(); i++ {
			result <- tuple{i, sliceValue.Index(i).Interface()}
		}
	case reflect.Map:
		for _, key := range sliceValue.MapKeys() {
			result <- tuple{key, sliceValue.MapIndex(key).Interface()}
		}
	default:
		panic("unreachable")
	}

}

func IsIterable(obj any) bool {
	value := reflect.ValueOf(obj)
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array || value.Kind() == reflect.Map
}

// IsSubset checks whether _a_ is a subset of _B_
func IsSubset[T comparable](a, B []T) bool {
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

func (s *Stack[T]) Push(v T) int {
	*s = append(*s, v)
	return len(*s) - 1
}

func (s *Stack[T]) Peek() T {
	return (*s)[len(*s)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(*s) == 0
}

func (s *Stack[T]) Pop() T {
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

type Queue[T any] []T

func (q *Queue[T]) Push(v T) int {
	*q = append(*q, v)
	return len(*q) - 1
}

func (q *Queue[T]) Peek() T {
	return (*q)[0]
}

func (q *Queue[T]) IsEmpty() bool {
	return len(*q) == 0
}

func (q *Queue[T]) Pop() T {
	v := (*q)[0]
	*q = (*q)[1:]
	return v
}

type Set[T comparable] []T

func (s *Set[T]) Add(v T) {
	if !s.Contains(v) {
		*s = append(*s, v)
	}
}

func (s *Set[T]) Remove(v T) {
	*s = slices.DeleteFunc(*s, func(t T) bool {
		return t == v
	})
}

func (s *Set[T]) Contains(v T) bool {
	return slices.Contains(*s, v)
}

func (s *Set[T]) Combine(other Set[T]) {
	for _, v := range other {
		s.Add(v)
	}
}

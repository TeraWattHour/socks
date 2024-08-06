package helpers

func Combine[T comparable, B any](a, b map[T]B) map[T]B {
	result := map[T]B{}
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

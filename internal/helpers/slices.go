package helpers

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

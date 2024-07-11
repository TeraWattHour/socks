package expression

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

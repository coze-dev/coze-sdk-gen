package util

func Choose[T any](cond bool, a, b T) T {
	if cond {
		return a
	}

	return b
}

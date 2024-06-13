package slice

func Map[T, U any](s []T, f func(T) U) []U {
	var result []U
	for _, v := range s {
		result = append(result, f(v))
	}
	return result
}

func Has[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func Filter[T any](s []T, f func(T) bool) []T {
	var result []T
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

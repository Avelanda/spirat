package utils

func Map[T, V any](ts []T, f func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = f(t)
	}
	return result
}

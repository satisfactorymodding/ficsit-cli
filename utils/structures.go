package utils

func CopyMap[T comparable, M any](m map[T]M) map[T]M {
	m2 := make(map[T]M, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

package utils

func Ifs[T any](a bool, b, c T) T {
	if a {
		return b
	}
	return c
}

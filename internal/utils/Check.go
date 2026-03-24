package utils

func Lengthcheck(min int, s ...string) bool {
	for _, v := range s {
		if len(v) < min {
			return false
		}
	}
	return true
}

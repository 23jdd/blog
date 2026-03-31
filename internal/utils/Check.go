package utils

import "regexp"

func Lengthcheck(min int, s ...string) bool {
	for _, v := range s {
		if len(v) < min {
			return false
		}
	}
	return true
}

var emailRegex *regexp.Regexp

func init() {
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
}
func EmailCheck(email string) bool {
	return emailRegex.MatchString(email)
}

package utils

import (
	"math/rand"
	"time"
)

func GenerateRandomCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = byte(r.Intn(10) + '0')
	}
	return string(code)
}

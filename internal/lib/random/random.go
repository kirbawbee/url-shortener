package random

import (
	"math/rand"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var rnd *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewRandomString(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}
	return string(b)
}

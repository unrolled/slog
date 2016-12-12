package slog

import (
	"math/rand"
	"time"
)

const (
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength = 62

	tokenLength = 36
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Token is an interface for generating tokens for the middleware.
type Token interface {
	Generate() string
}

type genericToken struct{}

// Generate returns a random string of the given length.
func (genericToken) Generate() string {
	b := make([]byte, tokenLength)
	for i := range b {
		b[i] = charset[rand.Intn(charsetLength)]
	}
	return string(b)
}

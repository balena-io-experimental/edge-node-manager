package proxyvisor

import (
	"math/rand"
	"time"
)

func Provision() (string, error) {
	// Simulate proxyvisor whilst we wait for it to be released by returning a random 62 char string
	return random(62), nil
}

func random(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

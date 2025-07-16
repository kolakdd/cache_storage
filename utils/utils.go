package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func GenerateToken(length int) string {
	bytes := generateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(bytes)
}

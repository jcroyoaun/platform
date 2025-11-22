package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func New() string {
	return strings.ToLower(rand.Text())
}

func Hash(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}

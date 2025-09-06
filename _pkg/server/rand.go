package server

import (
	"crypto/rand"
	"encoding/hex"
)

// RandID generates a random hex string of length 2n
func RandID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

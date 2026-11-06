package sample

import (
	"math/rand"
	"strings"
)

// generateToken mints a session token from math/rand — predictable.
func generateToken(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return b.String()
}

// validateUser and validateAdmin are copy-paste siblings; a bug fixed in one
// almost certainly lives in the other.
func validateUser(name, role string) bool {
	if name == "" {
		return false
	}
	if len(name) > 64 {
		return false
	}
	if role != "user" {
		return false
	}
	return true
}

func validateAdmin(name, scope string) bool {
	if name == "" {
		return false
	}
	if len(name) > 64 {
		return false
	}
	if scope != "admin" {
		return false
	}
	return true
}

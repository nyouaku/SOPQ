// ===== FILE: backend/pkg/utils/password.go =====
package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword menghasilkan bcrypt hash dari password plaintext
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword membandingkan password plaintext dengan hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashToken menghasilkan SHA256 hash untuk token yang panjang (>72 bytes)
// Digunakan untuk refresh token yang terlalu panjang untuk bcrypt
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// VerifyToken membandingkan token plaintext dengan hash SHA256.
// Perbandingan konstan-waktu: `==` pada string berhenti di byte pertama yang
// beda, sehingga waktu eksekusinya membocorkan berapa banyak prefiks yang cocok.
func VerifyToken(hashedToken, token string) bool {
	return subtle.ConstantTimeCompare([]byte(hashedToken), []byte(HashToken(token))) == 1
}
// ===== END FILE =====

// ===== FILE: backend/pkg/utils/jwt.go =====
package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	RoleID   int       `json:"role_id"`
	RoleName string    `json:"role_name"`
	jwt.RegisteredClaims
}

// GenerateAccessToken menghasilkan JWT access token
func GenerateAccessToken(userID uuid.UUID, username string, roleID int, roleName string, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken menghasilkan JWT refresh token (hanya user_id + jti).
// jti acak wajib: tanpa itu dua token untuk user yang sama di detik yang sama
// menghasilkan string identik, sehingga rotasi tidak pernah mengganti token.
func GenerateRefreshToken(userID uuid.UUID, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate token id: %w", err)
	}
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ID:        jti.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAccessToken mem-parsing dan memvalidasi access token
func ParseAccessToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ParseRefreshToken mem-parsing dan memvalidasi refresh token
func ParseRefreshToken(tokenString string, secret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user_id in token: %w", err)
		}
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("invalid token")
}

// ===== END FILE =====

package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type Claims struct {
	UserID  uint   `json:"user_id"`
	Login   string `json:"login"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-me"
	}
	return []byte(s)
}
func Generate(userID uint, login string, isAdmin bool) (string, error) {
	claims := Claims{userID, login, isAdmin, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret())
}
func Parse(token string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) { return secret(), nil })
	if err != nil || !t.Valid {
		return nil, err
	}
	return claims, nil
}

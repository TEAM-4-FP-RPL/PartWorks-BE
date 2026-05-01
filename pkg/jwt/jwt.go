package jwt

import (
	"errors"
	"os"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, email string, role string, ttl time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	claims := gjwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(ttl).Unix(),
	}
	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr string) (map[string]any, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	p := func(token *gjwt.Token) (any, error) {
		return []byte(secret), nil
	}
	parsed, err := gjwt.Parse(tokenStr, p)
	if err != nil {
		return nil, err
	}
	if claims, ok := parsed.Claims.(gjwt.MapClaims); ok && parsed.Valid {
		out := make(map[string]any)
		for k, v := range claims {
			out[k] = v
		}
		return out, nil
	}
	return nil, errors.New("invalid token")
}

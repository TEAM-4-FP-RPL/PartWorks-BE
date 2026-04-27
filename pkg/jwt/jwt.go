package jwt

import (
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

func VerifyToken(tokenStr string) (gjwt.MapClaims, error) {
        secret := os.Getenv("JWT_SECRET")
        if secret == "" {
                secret = "dev-secret"
        }
        token, err := gjwt.Parse(tokenStr, func(token *gjwt.Token) (interface{}, error) {
                return []byte(secret), nil
        })
        if err != nil {
                return nil, err
        }
        if claims, ok := token.Claims.(gjwt.MapClaims); ok && token.Valid {
                return claims, nil
        }
        return nil, gjwt.ErrSignatureInvalid
}

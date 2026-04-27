package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/jwt"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/pkg/response"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := jwt.VerifyToken(parts[1])
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		userID, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Roles middleware to restrict access based on user role
func Roles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				response.Error(w, http.StatusForbidden, "forbidden")
				return
			}

			authorized := false
			for _, role := range roles {
				if userRole == role {
					authorized = true
					break
				}
			}

			if !authorized {
				response.Error(w, http.StatusForbidden, "forbidden: insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

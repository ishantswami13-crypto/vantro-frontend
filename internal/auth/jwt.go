package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

var jwtSecret = []byte("vantro_super_secret_change_me")

// Middleware is a net/http middleware for JWT-protected endpoints.
// It is currently unused in the Fiber stack but kept for compatibility with simple http mux flows.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "missing auth token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(h, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		rawUID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "user_id missing", http.StatusUnauthorized)
			return
		}

		uid, err := uuid.Parse(rawUID)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user not authenticated")
	}
	return uid, nil
}

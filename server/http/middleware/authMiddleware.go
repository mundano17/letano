package middleware

import (
	"context"
	"errors"
	"net/http"
	"tano/internal/auth"
)

var ErrInvalidRequest = errors.New("invalid authorization header")
var ErrInvalidToken = errors.New("invalid token")

type userIDKey struct{}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := auth.VerifyToken(cookie.Value)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			nctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(nctx))
		},
	)
}

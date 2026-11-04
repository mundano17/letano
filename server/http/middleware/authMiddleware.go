package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func jwtKey() []byte {
	return []byte(os.Getenv("JWT_PVT_KEY"))
}

func VerifyToken(tokenString string) (string, error) {
	tok, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey(), nil
	})
	if err != nil {
		return "", err
	}
	sub, err := tok.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	return sub, err
}

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

			userID, err := VerifyToken(cookie.Value)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			nctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(nctx))
		},
	)
}

var ErrUserIDNotFoundInContext error = errors.New("User ID not found in context")

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return "", ErrUserIDNotFoundInContext
	}
	return userID, nil
}

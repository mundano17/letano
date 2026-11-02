package auth

import (
	"errors"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func jwtKey() []byte {
	return []byte(os.Getenv("JWT_PVT_KEY"))
}

func createAccessToken(userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iss": "tano_server",
			"sub": userID,
		})
	s, err := t.SignedString(jwtKey())
	return s, err
}

func createRefreshToken(userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
			"iss": "tanoserver",
			"sub": userID,
		})
	s, err := t.SignedString(jwtKey())
	return s, err
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

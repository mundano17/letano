package auth

import (
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
			"exp": time.Now().Add(5 * time.Minute).Unix(),
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
			"exp": time.Now().Add(2 * 24 * time.Hour).Unix(),
			"iss": "tanoserver",
			"sub": userID,
		})
	s, err := t.SignedString(jwtKey())
	return s, err
}

package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"tano/http/middleware"
	db "tano/internal/db/generated"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials  error = errors.New("invalid credentials")
	ErrMismatchedToken     error = errors.New("token mismatched")
	ErrWeakPassword        error = errors.New("invalid pasword")
	ErrUserAlreadyExists   error = errors.New("user already exists")
	ErrInvalidRefreshToken error = errors.New("invalid refresh token")
	ErrUserDoesNotExist    error = errors.New("user doesn't exist")
)

type authService struct {
	query *db.Queries
}

func newAuthService(query *db.Queries) *authService {
	return &authService{query}
}

type ErrPasswordValidator struct {
	Message string
}

func (e *ErrPasswordValidator) Error() string {
	return e.Message
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &ErrPasswordValidator{Message: "length of password must be greater than or equal to 8 characters"}
	}
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:,.<>?/", r):
			hasSpecial = true
		}
	}
	switch {
	case !hasUpper:
		return &ErrPasswordValidator{Message: "password must have a upper character"}
	case !hasLower:
		return &ErrPasswordValidator{Message: "password must have a lower character"}
	case !hasDigit:
		return &ErrPasswordValidator{Message: "password must have digits"}
	case !hasSpecial:
		return &ErrPasswordValidator{Message: "password must have a special character"}
	}
	return nil
}

func (service *authService) RegisterUserService(ctx context.Context, email string, password string) error {
	userExists := true
	_, err := service.query.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		userExists = false
	}

	if userExists {
		return ErrUserAlreadyExists
	}

	err = ValidatePassword(password)
	if err != nil {
		return err
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	_, err = service.query.CreateUser(ctx, db.CreateUserParams{UserID: uuid.NewString(), Email: email, PasswordHash: string(hash)})
	return err
}

func hashToken(tokenstring string) string {
	hash := sha256.Sum256([]byte(tokenstring))
	hashString := fmt.Sprintf("%x", hash)
	return hashString
}

func (service *authService) LoginUserService(ctx context.Context, email string, password string) (accessTokenString string, refreshTokenString string, err error) {
	res, err := service.query.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", ErrUserDoesNotExist
	}

	match, err := argon2id.ComparePasswordAndHash(password, res.PasswordHash)
	if err != nil {
		return "", "", err
	}
	if !match {
		return "", "", ErrInvalidCredentials
	}

	accessTokenString, err = createAccessToken(res.UserID)
	if err != nil {
		return "", "", err
	}

	refreshTokenString, err = createRefreshToken(res.UserID)
	if err != nil {
		return "", "", err
	}

	err = service.query.UpdateRefreshTokenHash(
		ctx,
		db.UpdateRefreshTokenHashParams{
			RefreshTokenHash: sql.NullString{
				String: hashToken(refreshTokenString),
				Valid:  true,
			},
			UserID: res.UserID,
		},
	)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	return accessTokenString, refreshTokenString, nil
}

func (service *authService) DeleteUserService(ctx context.Context, userID string, password string) error {
	res, err := service.query.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserDoesNotExist
		}
		return err
	}

	match, err := argon2id.ComparePasswordAndHash(password, res.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		return ErrInvalidCredentials
	}

	rows, err := service.query.DeleteUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserDoesNotExist
	}
	return nil
}

func (service *authService) RefreshTokenService(ctx context.Context, refreshTokenString string) (string, string, error) {
	userID, err := middleware.VerifyToken(refreshTokenString)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}

	hashedRefreshToken := hashToken(refreshTokenString)
	dbrefreshToken, err := service.query.GetRefreshTokenHashByUserID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if hashedRefreshToken != dbrefreshToken.String {
		_ = service.query.SetRefreshTokenHashToNULL(ctx, userID)
		return "", "", ErrMismatchedToken
	}

	newrefreshTokenString, err := createRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	newAccessTokenString, err := createAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	err = service.query.UpdateRefreshTokenHash(
		ctx,
		db.UpdateRefreshTokenHashParams{
			RefreshTokenHash: sql.NullString{
				String: hashToken(refreshTokenString),
				Valid:  true,
			},
			UserID: userID,
		},
	)
	if err != nil {
		return "", "", err
	}

	return newAccessTokenString, newrefreshTokenString, nil
}

func (service *authService) LogoutService(ctx context.Context, userID string) error {
	err := service.query.SetRefreshTokenHashToNULL(ctx, userID)
	return err
}

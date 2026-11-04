package auth

import (
	"errors"
	"net/http"
	"tano/http/middleware"
	db "tano/internal/db/generated"
	"tano/internal/utils"
)

func setTokenCookies(w http.ResponseWriter, path string, cookieName string, cookieValue string, age int) {
	http.SetCookie(w,
		&http.Cookie{
			Name:     cookieName,
			Value:    cookieValue,
			HttpOnly: true,
			Secure:   false, // SET TO TRUE IN PRODUCTION
			SameSite: http.SameSiteLaxMode,
			Path:     path,
			MaxAge:   age,
		},
	)
}

type AuthController struct {
	service *authService
}

func NewAuthController(query *db.Queries) *AuthController {
	return &AuthController{
		service: newAuthService(query),
	}
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AuthController) RegisterUserController(w http.ResponseWriter, r *http.Request) {
	req := utils.GetJson[RegisterUserRequest](w, r)
	if req == nil {
		return
	}

	err := c.service.RegisterUserService(r.Context(), req.Email, req.Password)
	if err != nil {
		if validateErr, ok := errors.AsType[*ErrPasswordValidator](err); ok {
			http.Error(w, validateErr.Message, http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrUserAlreadyExists) {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return

	}

	w.WriteHeader(http.StatusCreated)
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AuthController) LoginUserController(w http.ResponseWriter, r *http.Request) {
	req := utils.GetJson[LoginUserRequest](w, r)
	if req == nil {
		return
	}

	accessTokenString, refreshTokenString, err := c.service.LoginUserService(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ErrUserDoesNotExist) {
			http.Error(w, "User Doesn't Exist", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	setTokenCookies(w, "/", "access_token", accessTokenString, 15*60)
	setTokenCookies(w, "/auth/refresh", "refresh_token", refreshTokenString, 7*24*60*60)
	w.WriteHeader(http.StatusOK)
}

type DeleteUserRequest struct {
	Email    string
	Password string
}

func (c *AuthController) DeleteUserController(w http.ResponseWriter, r *http.Request) {
	req := utils.GetJson[DeleteUserRequest](w, r)
	if req == nil {
		return
	}

	ctx := r.Context()
	value, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = c.service.DeleteUserService(ctx, value, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ErrUserDoesNotExist) {
			http.Error(w, "User does not exist", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) RefreshTokenController(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	accessTokenString, refreshTokenString, err := c.service.RefreshTokenService(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			http.Error(w, "Invalid refresh token", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrMismatchedToken) {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	setTokenCookies(w, "/", "access_token", accessTokenString, 5*60)
	setTokenCookies(w, "/auth/refresh", "refresh_token", refreshTokenString, 2*24*60*60)
	w.WriteHeader(http.StatusOK)
}

// TODO Implement Logout Controller once the controller design is done.

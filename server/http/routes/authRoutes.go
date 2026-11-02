package routes

import (
	"net/http"
	"tano/http/middleware"
	"tano/internal/auth"
)

func NewAuthMux(controller *auth.AuthController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", controller.RegisterUserController)
	mux.HandleFunc("POST /login", controller.LoginUserController)
	mux.HandleFunc("POST /refresh", controller.RefreshTokenController)
	mux.Handle(
		"DELETE /delete",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.DeleteUserController)),
	)

	return mux
}

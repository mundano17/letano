package routes

import (
	"net/http"
	"tano/http/middleware"
	"tano/internal/upload"
)

func NewUploadMux(controller *upload.UploadController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(
		"POST /init",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.InitializeUploadController)),
	)
	mux.Handle(
		"POST /finish/{uploadID}",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.FinishUpload)),
	)
	mux.Handle(
		"POST /add",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.AddChunk)),
	)
	mux.Handle(
		"GET /status/{uploadID}",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.GetStatusController)),
	)
	mux.Handle(
		"POST /cancel/{uploadID}",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.CancelUploadController)),
	)
	mux.Handle(
		"POST /pause/{uploadID}",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.PauseUploadController)),
	)
	mux.Handle(
		"POST /resume/{uploadID}",
		middleware.AuthenticationMiddleware(http.HandlerFunc(controller.ResumeUploadController)),
	)

	return mux
}

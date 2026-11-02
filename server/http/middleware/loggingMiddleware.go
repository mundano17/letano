package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return (http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler := slog.NewJSONHandler(os.Stdout, nil)
			logger := slog.New(handler)
			start := time.Now()
			ctx := r.Context()

			logger.LogAttrs(
				ctx,
				slog.LevelInfo,
				"Request Started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			rw := &responseWriter{w, http.StatusAccepted}

			next.ServeHTTP(rw, r)

			logger.LogAttrs(
				ctx,
				slog.LevelInfo,
				"Request Completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status_code", rw.statusCode),
				slog.Int64("duration_us", time.Since(start).Microseconds()),
			)
		},
	))
}

package middleware

import (
	"net/http"
	"slices"
	"strings"
)

func isPreflight(r *http.Request) bool {
	return r.Method == "OPTIONS" &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get("Access-Control-Request-Method") != "" &&
		r.Header.Get("Access-Control-Request-Headers") != ""

}

var methodAllowList = []string{"GET", "POST", "DELETE", "OPTIONS"}
var headersAllowList = []string{"content-type", "authorization"}
var originAllowList = []string{"http://127.0.0.1:3000", "http://localhost:3000"}

func isValidPreflightHeaders(headers []string) bool {
	for _, header := range headers {
		header = strings.TrimSpace(header)
		header = strings.ToLower(header)
		if !slices.Contains(headersAllowList, header) {
			return false
		}
	}
	return true
}

func preFlightHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	method := r.Header.Get("Access-Control-Request-Method")
	headers := strings.Split(
		r.Header.Get("Access-Control-Request-Headers"),
		",",
	)
	if slices.Contains(originAllowList, origin) && slices.Contains(methodAllowList, method) && isValidPreflightHeaders(headers) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methodAllowList, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(headersAllowList, ", "))
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if isPreflight(r) {
				preFlightHandler(w, r)
				w.Header().Add("Vary", "Origin")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			origin := r.Header.Get("Origin")
			if slices.Contains(originAllowList, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Add("Vary", "Origin")
			next.ServeHTTP(w, r)
		},
	)
}

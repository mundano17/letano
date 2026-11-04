package utils

import (
	"encoding/json"
	"net/http"
)

func GetJson[T any](w http.ResponseWriter, r *http.Request) *T {
	obj := new(T)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return nil
	}
	return obj
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

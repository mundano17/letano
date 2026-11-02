package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var ErrUserIDNotFoundInContext error = errors.New("User ID not found in context")

type userIDKey struct {
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	if !ok {
		return "", ErrUserIDNotFoundInContext
	}
	return userID, nil
}

func GetJson[T any](w http.ResponseWriter, r *http.Request) *T {
	obj := new(T)
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return nil
	}
	return obj
}

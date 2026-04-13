package utils

import (
	"net/http"
)

func GetUserIDFromContext(r *http.Request) (uint64, error) {
	val := r.Context().Value("user_id")
	if val == nil {
		return 0, ErrUnauthorized
	}
	userID, ok := val.(uint64)
	if !ok {
		return 0, ErrUnauthorized
	}
	return userID, nil
}

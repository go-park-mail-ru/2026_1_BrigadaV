package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/csrf"
)

type CSRFHandler struct{}

func NewCSRFHandler() *CSRFHandler {
	return &CSRFHandler{}
}

// GetToken возвращает CSRF токен для текущего запроса
// @Summary Get CSRF token
// @Description Returns CSRF token for the current session
// @Tags security
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/csrf-token [get]
func (h *CSRFHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"csrf_token": token,
	})
}
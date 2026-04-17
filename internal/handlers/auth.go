package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Login:    req.Login,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		status := http.StatusBadRequest
		resp := map[string]string{"error": err.Error()}
		switch err.Error() {
		case "login already exists":
			status = http.StatusConflict
			resp["field"] = "login"
		case "nickname already exists":
			status = http.StatusConflict
			resp["field"] = "nickname"
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "user created",
		"user_id": user.ID,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}
	user, token, err := h.authService.Login(r.Context(), service.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	json.NewEncoder(w).Encode(dto.LoginResponse{
		UserID:    user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "logout failed"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	json.NewEncoder(w).Encode(dto.LoginResponse{
		UserID:    user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	})
}

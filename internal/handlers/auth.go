package handlers

import (
	"encoding/json"
	"guidely-app/internal/dto"
	"guidely-app/internal/service"
	"guidely-app/internal/utils"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Nickname: req.Nickname,
		Password: req.Password,
	})
	if err != nil {
		if err.Error() == "nickname already exists" {
			utils.WriteJSONErrorWithField(w, err, "nickname", http.StatusConflict)
			return
		}
		if err.Error() == "nickname must be at least 3 characters and max 50" {
			utils.WriteJSONErrorWithField(w, err, "nickname", http.StatusBadRequest)
			return
		}
		if err.Error() == "password must be at least 8 characters" {
			utils.WriteJSONErrorWithField(w, err, "password", http.StatusBadRequest)
			return
		}
		utils.WriteJSONError(w, err, http.StatusBadRequest)
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
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	user, token, err := h.authService.Login(r.Context(), service.LoginInput{
		Nickname: req.Nickname,
		Password: req.Password,
	})
	if err != nil {
		utils.WriteJSONError(w, utils.ErrInvalidCredentials, http.StatusUnauthorized)
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
		utils.WriteJSONError(w, utils.ErrUnauthorized, http.StatusUnauthorized)
		return
	}
	if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
		utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
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
	userID, err := utils.GetUserIDFromContext(r)
	if err != nil {
		utils.WriteJSONError(w, err, http.StatusUnauthorized)
		return
	}
	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		utils.WriteJSONError(w, utils.ErrNotFound, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(dto.LoginResponse{
		UserID:    user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	})
}

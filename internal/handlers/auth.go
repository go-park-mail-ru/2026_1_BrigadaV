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
	user, err := h.authService.Register(r.Context(), service.RegisterInput{
		Login:    req.Login,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		switch err {
		case utils.ErrBadRequest:
			utils.WriteJSONError(w, err, http.StatusBadRequest)
		case utils.ErrConflict:
			utils.WriteJSONError(w, err, http.StatusConflict)
		default:
			utils.WriteJSONError(w, utils.ErrInternal, http.StatusInternalServerError)
		}
		return
	}
	utils.WriteJSON(w, map[string]interface{}{
		"message": "user created",
		"user_id": user.ID,
	}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, utils.ErrBadRequest, http.StatusBadRequest)
		return
	}
	user, token, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		utils.WriteJSONError(w, utils.ErrInvalidCredentials, http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(utils.SessionCookieExpiry),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	utils.WriteJSON(w, dto.LoginResponse{
		UserID:    user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}, http.StatusOK)
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
	utils.WriteJSON(w, map[string]string{"message": "logged out"}, http.StatusOK)
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
	utils.WriteJSON(w, dto.LoginResponse{
		UserID:    user.ID,
		Login:     user.Login,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}, http.StatusOK)
}

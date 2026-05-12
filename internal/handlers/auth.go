package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"guidely-app/internal/logger"
	pb "guidely-app/pkg/pb/auth"

	"github.com/gorilla/csrf"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	client pb.AuthServiceClient
}

func NewAuthHandler(client pb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: client}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := h.client.Register(r.Context(), &pb.RegisterRequest{
		Login:    req.Login,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		logger.Error(r.Context(), "register gRPC error", logrus.Fields{"error": err, "login": req.Login})
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			http.Error(w, st.Message(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	loginResp, err := h.client.Login(r.Context(), &pb.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		logger.Warn(r.Context(), "auto-login after register failed", logrus.Fields{"error": err, "login": req.Login})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "user created",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    loginResp.Token,
		MaxAge:   7 * 24 * 60 * 60,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Отдаём CSRF-токен сразу после регистрации, чтобы фронтенд
	// мог использовать его в следующем запросе без отдельного GET /api/csrf-token
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  loginResp.UserId,
		"nickname": loginResp.Nickname,
		"message":  "user created",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Login(r.Context(), &pb.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		logger.Error(r.Context(), "login gRPC error", logrus.Fields{"error": err, "login": req.Login})
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    resp.Token,
		MaxAge:   7 * 24 * 60 * 60,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	// Отдаём CSRF-токен после логина
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  resp.UserId,
		"nickname": resp.Nickname,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := h.client.Logout(r.Context(), &pb.LogoutRequest{Token: cookie.Value}); err != nil {
		logger.Error(r.Context(), "logout gRPC error", logrus.Fields{"error": err})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	resp, err := h.client.GetUser(r.Context(), &pb.GetUserRequest{UserId: userID})
	if err != nil {
		logger.Error(r.Context(), "get user gRPC error", logrus.Fields{"error": err, "user_id": userID})
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

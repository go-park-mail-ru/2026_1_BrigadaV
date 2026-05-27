package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"guidely-app/internal/dto"
	"guidely-app/internal/logger"
	"guidely-app/pkg/config"
	pb "guidely-app/pkg/pb/auth"

	"github.com/gorilla/csrf"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	client pb.AuthServiceClient
	cfg    *config.Config
}

func NewAuthHandler(client pb.AuthServiceClient, cfg *config.Config) *AuthHandler {
	return &AuthHandler{client: client, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request body"}`))
		return
	}

	_, err := h.client.Register(r.Context(), &pb.RegisterRequest{
		Login:    req.Login,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		logger.Error(r.Context(), "register gRPC error", logrus.Fields{"error": err, "login": req.Login})
		w.Header().Set("Content-Type", "application/json")
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"` + st.Message() + `"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
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
		w.Write([]byte(`{"message":"user created"}`))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    loginResp.Token,
		MaxAge:   7 * 24 * 60 * 60,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   h.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response := dto.LoginResponse{
		UserID:    loginResp.UserId,
		Nickname:  loginResp.Nickname,
		AvatarURL: loginResp.AvatarUrl,
	}
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := easyjson.MarshalToWriter(&response, w); err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request body"}`))
		return
	}

	resp, err := h.client.Login(r.Context(), &pb.LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		logger.Error(r.Context(), "login gRPC error", logrus.Fields{"error": err, "login": req.Login})
		w.Header().Set("Content-Type", "application/json")
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid credentials"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    resp.Token,
		MaxAge:   7 * 24 * 60 * 60,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   h.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response := dto.LoginResponse{
		UserID:    resp.UserId,
		Nickname:  resp.Nickname,
		AvatarURL: resp.AvatarUrl,
	}
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("Content-Type", "application/json")
	if _, err := easyjson.MarshalToWriter(&response, w); err != nil {
		logger.Error(r.Context(), "easyjson marshal error", logrus.Fields{"error": err})
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	if _, err := h.client.Logout(r.Context(), &pb.LogoutRequest{Token: cookie.Value}); err != nil {
		logger.Error(r.Context(), "logout gRPC error", logrus.Fields{"error": err})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint64)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	resp, err := h.client.GetUser(r.Context(), &pb.GetUserRequest{UserId: userID})
	if err != nil {
		logger.Error(r.Context(), "get user gRPC error", logrus.Fields{"error": err, "user_id": userID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"user not found"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(r.Context(), "json encode error", logrus.Fields{"error": err})
	}
}

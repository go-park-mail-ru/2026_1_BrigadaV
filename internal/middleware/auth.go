package middleware

import (
	"context"
	"encoding/json"
	"guidely-app/internal/repository"
	"net/http"
	"time"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	sessionRepo repository.SessionRepository
	userRepo    *repository.UserRepo
}

func NewAuthMiddleware(sessionRepo repository.SessionRepository, userRepo *repository.UserRepo) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		userRepo: userRepo,
	}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "missing session cookie"})
			return
		}
		session, err := m.sessionRepo.GetByToken(r.Context(), cookie.Value)
		if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "invalid or expired session"})
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserIDFromContext(r *http.Request) int64 {
	if userID, ok := r.Context().Value(UserIDKey).(int64); ok {
		return userID
	}
	return 0
}

func (m *AuthMiddleware) AuthenticateAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "missing session cookie"})
			return
		}
		
		session, err := m.sessionRepo.GetByToken(r.Context(), cookie.Value)
		if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized", "message": "invalid or expired session"})
			return
		}
		
		user, err := m.userRepo.GetByID(r.Context(), session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
			return
		}
		
		if user.Role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "forbidden", "message": "admin access required"})
			return
		}
		
		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		ctx = context.WithValue(ctx, "role", user.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
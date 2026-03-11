package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticateMiddleware(t *testing.T) {
	resetGlobals()
	h := newTestHandlers()

	token, _ := generateSessionToken()
	userID := uint64(42)
	h.sessions[token] = Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		uid := r.Context().Value("user_id")
		assert.Equal(t, userID, uid, "user_id in context should match session")
		w.WriteHeader(http.StatusOK)
	})

	handler := h.authenticate(next)

	t.Run("no cookie", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.False(t, nextCalled)
	})

	t.Run("valid cookie", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, nextCalled)
	})

	t.Run("expired session", func(t *testing.T) {
		expiredToken, _ := generateSessionToken()
		h.sessions[expiredToken] = Session{
			Token:     expiredToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		nextCalled = false
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: expiredToken})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.False(t, nextCalled)
	})

	t.Run("unknown token", func(t *testing.T) {
		unknownToken, _ := generateSessionToken()
		nextCalled = false
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: unknownToken})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.False(t, nextCalled)
	})
}

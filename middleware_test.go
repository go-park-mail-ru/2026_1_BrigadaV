package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		if uid := r.Context().Value("user_id"); uid != userID {
			t.Errorf("context user_id = %v, want %d", uid, userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := h.authenticate(next)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no cookie: expected 401, got %d", resp.StatusCode)
	}
	if nextCalled {
		t.Error("next handler called when no cookie")
	}
	nextCalled = false

	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid cookie: expected 200, got %d", resp.StatusCode)
	}
	if !nextCalled {
		t.Error("next handler not called")
	}
	nextCalled = false

	expiredToken, _ := generateSessionToken()
	h.sessions[expiredToken] = Session{
		Token:     expiredToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: expiredToken})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expired session: expected 401, got %d", resp.StatusCode)
	}
	if nextCalled {
		t.Error("next handler called for expired session")
	}
	nextCalled = false

	unknownToken, _ := generateSessionToken()
	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: unknownToken})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unknown token: expected 401, got %d", resp.StatusCode)
	}
}

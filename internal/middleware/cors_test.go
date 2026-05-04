package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS("http://guidely.ru", "https://guidely.ru")(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://guidely.ru")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://guidely.ru", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS("http://guidely.ru")(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Preflight(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS("http://guidely.ru")(handler)

	req := httptest.NewRequest("OPTIONS", "/api/login", nil)
	req.Header.Set("Origin", "http://guidely.ru")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://guidely.ru", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestCORS_MultipleOrigins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS("http://guidely.ru", "https://guidely.ru", "http://localhost:3000")(handler)

	for _, origin := range []string{"http://guidely.ru", "https://guidely.ru", "http://localhost:3000"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"), "origin: %s", origin)
	}
}

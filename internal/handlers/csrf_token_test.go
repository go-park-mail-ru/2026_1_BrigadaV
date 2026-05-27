package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSRFHandler_GetToken(t *testing.T) {
	handler := NewCSRFHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/csrf-token", nil)
	w := httptest.NewRecorder()

	handler.GetToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "csrf_token")
}

func TestCSRFHandler_GetToken_MultipleCalls(t *testing.T) {
	handler := NewCSRFHandler()
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/csrf-token", nil)
		w := httptest.NewRecorder()
		handler.GetToken(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

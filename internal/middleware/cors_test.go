package middleware

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestCORS(t *testing.T) {
// 	allowedOrigin := "http://example.com"
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// 	})

// 	middleware := CORS(allowedOrigin)
// 	wrappedHandler := middleware(handler)

// 	tests := []struct {
// 		name            string
// 		method          string
// 		origin          string
// 		expectedStatus  int
// 		expectedHeaders map[string]string
// 	}{
// 		{
// 			name:           "OPTIONS request",
// 			method:         http.MethodOptions,
// 			origin:         allowedOrigin,
// 			expectedStatus: http.StatusOK,
// 			expectedHeaders: map[string]string{
// 				"Access-Control-Allow-Origin":  allowedOrigin,
// 				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
// 				"Access-Control-Allow-Headers": "Content-Type, Authorization",
// 			},
// 		},
// 		{
// 			name:           "GET request",
// 			method:         http.MethodGet,
// 			origin:         allowedOrigin,
// 			expectedStatus: http.StatusOK,
// 			expectedHeaders: map[string]string{
// 				"Access-Control-Allow-Origin": allowedOrigin,
// 			},
// 		},
// 		{
// 			name:           "POST request with different origin",
// 			method:         http.MethodPost,
// 			origin:         "http://other.com",
// 			expectedStatus: http.StatusOK,
// 			expectedHeaders: map[string]string{
// 				"Access-Control-Allow-Origin": allowedOrigin,
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := httptest.NewRequest(tt.method, "/", nil)
// 			req.Header.Set("Origin", tt.origin)
// 			rec := httptest.NewRecorder()

// 			wrappedHandler.ServeHTTP(rec, req)

// 			assert.Equal(t, tt.expectedStatus, rec.Code)

// 			for header, expectedValue := range tt.expectedHeaders {
// 				assert.Equal(t, expectedValue, rec.Header().Get(header))
// 			}

// 			if tt.method != http.MethodOptions {
// 				assert.Equal(t, "OK", rec.Body.String())
// 			}
// 		})
// 	}
// }

package middleware

import (
    "log"
    "net/http"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            log.Printf("[CORS] Request Origin: %s", origin)
            log.Printf("[CORS] Allowed origins: %v", allowedOrigins)

            allowed := false
            for _, ao := range allowedOrigins {
                if origin == ao {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                log.Printf("[CORS] Allowed origin: %s", origin)
            } else if origin != "" {
                log.Printf("[CORS] Rejected origin: %s", origin)
            }

            w.Header().Set("Access-Control-Allow-Credentials", "true")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

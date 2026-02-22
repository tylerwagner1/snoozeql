package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"snoozeql/internal/config"
)

// RequestIDKey is the context key for request ID
type RequestIDKey struct{}

// RealIPKey is the context key for real IP
type RealIPKey struct{}

// NewRequestID adds a request ID to the context
func NewRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = generateReqID()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey{}, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TrackRequestDuration tracks request duration
func TrackRequestDuration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// APIKeyAuth validates API keys from the Authorization header
func APIKeyAuth(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract API key (format: "Bearer <key>" or just "<key>")
			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			apiKey = strings.TrimSpace(apiKey)

			// Development mode: allow API key "dev-key" without validation
			if apiKey == "dev-key" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate key against database (to be implemented in phase 3)
			_ = cfg

			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds CORS headers to responses
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// generateReqID generates a unique request ID
func generateReqID() string {
	return "req-" + time.Now().UTC().Format("20060102150405")
}

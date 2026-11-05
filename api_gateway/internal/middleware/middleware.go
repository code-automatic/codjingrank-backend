package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"innogen-backend/shared/response"
)

// CORS returns middleware that allows specified origins.
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := make(map[string]bool)
	for _, o := range strings.Split(allowedOrigins, ",") {
		origins[strings.TrimSpace(o)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if r.Method == http.MethodOptions {
				if origins[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-User-Email, X-User-Role")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if origins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID ensures every request has an X-Request-ID header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ctxKey string

const ctxKeyRequestID ctxKey = "requestId"

// GetRequestID returns the request ID from the request context.
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(ctxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// RequestLogger logs every request.
func RequestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("clientIP", clientIP),
			slog.String("requestId", GetRequestID(r)),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// Recover catches panics and returns 500.
func Recover(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered",
					slog.String("requestId", GetRequestID(r)),
					slog.Any("panic", rec),
				)
				response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// BodyLimit limits request body size.
func BodyLimit(maxBytes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter provides in-memory IP-based rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(limit int, windowSeconds int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Duration(windowSeconds) * time.Second,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for range time.Tick(time.Minute) {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, times := range rl.visitors {
			var recent []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.visitors, ip)
			} else {
				rl.visitors[ip] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if an IP is allowed to make a request.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	var recent []time.Time
	for _, t := range rl.visitors[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	rl.visitors[ip] = recent
	return len(recent) <= rl.limit
}

// RateLimit returns middleware that rate-limits requests using the given limiter.
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if !rl.Allow(ip) {
				response.ErrorSimple(w, http.StatusTooManyRequests, "Too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

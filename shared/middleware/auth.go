package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/response"
)

// contextKey is an unexported type to avoid context key collisions.
type contextKey string

const (
	UserIDKey    contextKey = "userId"
	UserEmailKey contextKey = "userEmail"
	UserRoleKey  contextKey = "userRole"
)

// Claims represents the JWT claims extracted from a token.
type Claims struct {
	jwtlib.RegisteredClaims
	UserID int    `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// Auth returns an HTTP middleware that validates a Bearer JWT token and injects
// user claims into the request context. On failure, it writes a 401 JSON error
// response and stops the chain.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.ErrorSimple(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenString == "" {
				response.ErrorSimple(w, http.StatusUnauthorized, "Invalid token format")
				return
			}

			claims := &Claims{}
			token, err := jwtlib.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
					return nil, jwtlib.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.ErrorSimple(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(UserIDKey).(int)
	return id, ok
}

// GetUserEmail extracts the user email from the request context.
func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}

// XUserID returns middleware that extracts a user ID from the X-User-ID header.
// This is a temporary authentication mechanism for development.
// In production, use Auth() with JWT Bearer tokens instead.
func XUserID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := r.Header.Get("X-User-ID")
			if userIDStr == "" {
				response.ErrorSimple(w, http.StatusUnauthorized, "Missing X-User-ID header")
				return
			}

			userID, err := strconv.Atoi(userIDStr)
			if err != nil || userID <= 0 {
				response.ErrorSimple(w, http.StatusUnauthorized, "Invalid X-User-ID")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

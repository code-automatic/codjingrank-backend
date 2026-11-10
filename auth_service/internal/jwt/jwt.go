package jwt

import (
	"innogen-backend/shared/middleware"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	jwtlib.RegisteredClaims
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// Service handles JWT token generation and validation.
type Service struct {
	secret []byte
}

var (
	// ErrTokenExpired is returned when a token has expired.
	ErrTokenExpired = jwtlib.ErrTokenExpired
)

// NewService creates a new JWT Service with the given signing secret.
func NewService(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

// GenerateToken creates a signed JWT token for the given user with the specified TTL in minutes.
func (s *Service) GenerateToken(userID string, email, role string, ttlMinutes int) (string, error) {
	if !middleware.ValidUUID(userID) {
		return "", jwtlib.ErrTokenInvalidClaims
	}
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Duration(ttlMinutes) * time.Minute)),
			Subject:   userID,
		},
		UserID: userID,
		Email:  email,
		Role:   role,
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ParseToken parses and validates a JWT token string, returning the claims.
func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, jwtlib.ErrSignatureInvalid
		}
		return s.secret, nil
	}, jwtlib.WithValidMethods([]string{"HS256"}), jwtlib.WithExpirationRequired())
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || !middleware.ValidUUID(claims.UserID) || claims.Subject != claims.UserID {
		return nil, jwtlib.ErrSignatureInvalid
	}

	return claims, nil
}

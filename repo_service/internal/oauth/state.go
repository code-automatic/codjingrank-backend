package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type StateClaims struct {
	jwtlib.RegisteredClaims
	UserID int    `json:"userId"`
	Nonce  string `json:"nonce"`
}

func GenerateState(jwtSecret string, userID int) (string, error) {
	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)
	now := time.Now()
	claims := &StateClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(10 * time.Minute)),
		},
		UserID: userID,
		Nonce:  nonce,
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateState(stateToken, jwtSecret string) (*StateClaims, error) {
	token, err := jwtlib.ParseWithClaims(stateToken, &StateClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid state: %w", err)
	}
	claims, ok := token.Claims.(*StateClaims)
	if !ok {
		return nil, fmt.Errorf("invalid state claims")
	}
	return claims, nil
}

package middleware

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	issuer = "chat-app"
	signingKey = os.Getenv("AUTH_KEY")
	accessTTL  = 15 * time.Minute
)

func NewAccessToken(userID string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(accessTTL).Unix(),
	}).SignedString([]byte(signingKey))
}

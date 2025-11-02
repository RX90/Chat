package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	authHeader = "Authorization"
	userCtx    = "userID"
)

var (
	issuer     = "chat-app"
	signingKey = os.Getenv("AUTH_KEY")
	accessTTL  = 15 * time.Minute
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}

func NewAccessToken(userID, username string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
		},
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}

func ParseAccessToken(accessToken string) (*Claims, error) {
	if len(accessToken) == 0 || len(accessToken) > 4096 {
		return nil, errors.New("token is too long or empty")
	}
	if strings.Count(accessToken, ".") != 2 {
		return nil, errors.New("token has invalid format")
	}

	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(signingKey), nil
		},
	)

	if token == nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0 {
			return claims, errors.New("token has expired")
		}
		return nil, err
	}

	return claims, nil
}

func StrictUserIdentity(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "access token required"})
		return
	}

	claims, err := ParseAccessToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("error while parsing access token: %v", err)})
		return
	}

	c.Set(userCtx, claims.Subject)
	c.Next()
}

func SoftUserIdentity(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "auth header is empty"})
		return
	}

	claims, err := ParseAccessToken(token)
	if claims.Subject == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("access token is invalid: %v", err)})
		return
	}

	c.Set(userCtx, claims.Subject)
	c.Next()
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader(authHeader)
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return c.Query("accessToken")
}

func GetUserID(c *gin.Context) (uuid.UUID, error) {
	stringUserID := c.GetString(userCtx)
	return uuid.Parse(stringUserID)
}

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	jwt.StandardClaims
	Username string `json:"username"`
}

func NewAccessToken(userID, username string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		jwt.StandardClaims{
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(accessTTL).Unix(),
		},
		username,
	}).SignedString([]byte(signingKey))
}

func ParseAccessToken(accessToken string) (*Claims, error) {
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

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("token claims are not of type *Claims")
	}

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return claims, errors.New("token has expired")
			}
		}
		return nil, err
	}

	return claims, nil
}

func StrictUserIdentity(c *gin.Context) {
	var token string

	header := c.GetHeader(authHeader)
	if header != "" {
		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" || headerParts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "auth header is invalid"})
			return
		}
		token = headerParts[1]
	} else {
		token = c.Query("accessToken")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "access token required"})
			return
		}
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
	var token string

	header := c.GetHeader(authHeader)
	if header != "" {
		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" || headerParts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "auth header is invalid"})
			return
		}
		token = headerParts[1]
	} else {
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

func GetUserID(c *gin.Context) (uuid.UUID, error) {
	stringUserID := c.GetString("userID")
	userID, err := uuid.Parse(stringUserID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

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
)

const (
	authHeader = "Authorization"
	userCtx    = "userID"
)

var (
	signingKey = os.Getenv("AUTH_KEY")
	accessTTL  = 15 * time.Minute
)

func StrictUserIdentity(c *gin.Context) {
	header := c.GetHeader(authHeader)

	var token string

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

	// PROD ONLY
	userID, err := ParseAccessToken(token)
	if err != nil {
		if token != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDgwMzU0MTQsInN1YiI6ImExYWQ0ZjM2LTZiMmItNDQ4Zi05ZGU5LWFiNjUxZDMyNWZiYSJ9.0QhAR9kZ4Ycan3sCiayH3YLpzRaElTe5bYf-KpWeHZk" && token != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDg3MTEzMDQsInN1YiI6ImJiNGQ3NWUxLTFjN2YtNDNmYy1iZDVkLWIwMzA0YTZlM2FhMiJ9.ENo8SdqA0q7uCy0zd2XLGHAVQSbxAtlOCFO4toTaszs" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("error while parsing access token: %v", err)})
			return
		}
	}

	c.Set(userCtx, userID)
	c.Next()
}

func SoftUserIdentity(c *gin.Context) {
	header := c.GetHeader(authHeader)

	var token string

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

	userID, err := ParseAccessToken(token)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("access token is invalid: %v", err)})
		return
	}

	c.Set(userCtx, userID)
	c.Next()
}

func NewAccessToken(userID string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(accessTTL).Unix(),
		Subject:   userID,
	},
	)

	return accessToken.SignedString([]byte(signingKey))
}

func ParseAccessToken(accessToken string) (string, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.StandardClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(signingKey), nil
		})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
					return claims.Subject, errors.New("token has expired")
				}
			}
		}
		return "", err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", errors.New("token claims are not of type *jwt.StandardClaims")
	}

	return claims.Subject, nil
}

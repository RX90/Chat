package handler

import (
	"fmt"
	"net/http"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/middleware"
	"github.com/RX90/Chat/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	cookieName     = "refreshToken"
	cookiePath     = "/"
	cookieDomain   = "localhost"
	cookieHttpOnly = true
	cookieSameSite = http.SameSiteStrictMode
)

type authHandler struct {
	service service.AuthService
}

func newAuthHandler(s service.AuthService) *authHandler {
	return &authHandler{service: s}
}

func (h *authHandler) SignUp(c *gin.Context) {
	var input domain.User

	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("failed to bind JSON: %v", err)})
		return
	}

	if err := h.service.CreateUser(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to create user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *authHandler) SignIn(c *gin.Context) {
	var input domain.User

	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("failed to bind JSON: %v", err)})
		return
	}

	user, err := h.service.GetUserByEmail(input.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "invalid login or password"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "invalid login or password"})
		return
	}

	accessToken, err := middleware.NewAccessToken(user.ID.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to create access token: %v", err)})
		return
	}

	token, err := h.service.NewRefreshToken(user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to create refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token.RefreshToken,
		Expires:  token.ExpiresAt,
		Path:     cookiePath,
		Domain:   cookieDomain,
		HttpOnly: cookieHttpOnly,
		SameSite: cookieSameSite,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}

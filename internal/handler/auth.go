package handler

import (
	"fmt"
	"net/http"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/middleware"
	"github.com/RX90/Chat/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	cookieName     = "refreshToken"
	cookiePath     = "/"
	cookieHttpOnly = true
	cookieSameSite = http.SameSiteLaxMode
	cookieMaxAge   = -1
)

type authHandler struct {
	service service.AuthService
}

func newAuthHandler(s service.AuthService) *authHandler {
	return &authHandler{service: s}
}

func (h *authHandler) SignUp(c *gin.Context) {
	var input dto.SignUpUser

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
	var input dto.SignInUser

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

	accessToken, err := middleware.NewAccessToken(user.ID.String(), user.Username)
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
		HttpOnly: cookieHttpOnly,
		SameSite: cookieSameSite,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}

func (h *authHandler) SignOut(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("failed to get userID: %v", err)})
		return
	}

	if err := h.service.DeleteRefreshToken(userID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to delete refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   cookiePath,
		MaxAge: cookieMaxAge,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *authHandler) Refresh(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("failed to get userID: %v", err)})
		return
	}

	refreshToken, err := c.Cookie("refreshToken")
	if err != nil || refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "refresh token is missing"})
		return
	}

	if err := h.service.CheckRefreshToken(userID, refreshToken); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": fmt.Sprintf("refresh token is invalid: %v", err)})
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to get user info: %v", err)})
		return
	}

	accessToken, err := middleware.NewAccessToken(userID.String(), user.Username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create access token: %v", err)})
		return
	}

	token, err := h.service.NewRefreshToken(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't create refresh token: %v", err)})
		return
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token.RefreshToken,
		Expires:  token.ExpiresAt,
		Path:     cookiePath,
		HttpOnly: cookieHttpOnly,
		SameSite: cookieSameSite,
	}

	http.SetCookie(c.Writer, cookie)

	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}

func (h *authHandler) Verify(c *gin.Context) {
	c.Status(http.StatusOK)
}

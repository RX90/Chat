package handler

import (
	"fmt"
	"net/http"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/service"
	"github.com/gin-gonic/gin"
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

	if err := h.service.CreateUser(input); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to create user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

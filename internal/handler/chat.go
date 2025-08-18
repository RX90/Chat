package handler

import (
	"fmt"
	"net/http"

	"github.com/RX90/Chat/internal/server/ws"
	"github.com/RX90/Chat/internal/service"
	"github.com/gin-gonic/gin"
)

type chatHandler struct {
	service service.ChatService
}

func newChatHandler(s service.ChatService) *chatHandler {
	return &chatHandler{service: s}
}

func (h *chatHandler) ServeWS(c *gin.Context) {
	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't upgrade HTTP request: %v", err)})
		return
	}
	
	ws.ServeClient(conn, h.service)
}

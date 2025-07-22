package handler

import (
	"encoding/json"
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

	client := ws.NewClient(conn, h.service)

	msgs, err := h.service.GetMessages()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to load chat history: %v", err)})
		return
	} else {
		for _, msg := range *msgs {
			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to marshal history message: %v", err)})
				return
			}
			if err := client.SendMessage(jsonMsg); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to send message: %v", err)})
				return
			}
		}
	}
}

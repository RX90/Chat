package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RX90/Chat/internal/repo"
	"github.com/RX90/Chat/internal/ws"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	hub  *ws.Hub
	repo *repo.Repo
}

func NewHandler(hub *ws.Hub, repo *repo.Repo) *Handler {
	return &Handler{hub: hub, repo: repo}
}

func (h *Handler) HandleChatPage(c *gin.Context) {
	c.HTML(http.StatusOK, "chat.html", nil)
}

func (h *Handler) ServeWs(c *gin.Context) {
	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't upgrade HTTP request: %v", err)})
		return
	}

	client := ws.NewClient(h.hub, conn)

	go client.ReadPump(h.repo)
	go client.WritePump()

	msgs, err := h.repo.GetMessages()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to load chat history: %v", err)})
		return
	} else {
		for _, msg := range msgs {
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

package handler

import (
	"github.com/RX90/Chat/internal/service"
)

type Handler struct {
	Auth *authHandler
	Chat *chatHandler
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		Auth: newAuthHandler(service.Auth),
		Chat: newChatHandler(service.Chat),
	}
}

package service

import (
	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/RX90/Chat/internal/repo"
)

type chatService struct {
	repo repo.ChatRepo
}

func newChatService(r repo.ChatRepo) ChatService {
	return &chatService{repo: r}
}

func (s *chatService) CreateMessage(msg *entities.Message) (dto.OutgoingMessage, error) {
	return s.repo.CreateMessage(msg)
}

func (s *chatService) GetMessages() ([]dto.OutgoingMessage, error) {
	return s.repo.GetMessages()
}

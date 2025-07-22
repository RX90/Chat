package service

import (
	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/repo"
)

type chatService struct {
	repo repo.ChatRepo
}

func newChatService(r repo.ChatRepo) ChatService {
	return &chatService{repo: r}
}

func (s *chatService) CreateMessage(msg domain.Message) (*domain.Message, error) {
	return s.repo.CreateMessage(msg)
}

func (s *chatService) GetMessages() (*[]domain.Message, error) {
	return s.repo.GetMessages()
}
package service

import (
	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/repo"
	"github.com/google/uuid"
)

type Service struct {
	Auth AuthService
	Chat ChatService
}

type AuthService interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	NewRefreshToken(userID uuid.UUID) (*domain.Token, error)
}

type ChatService interface {
	CreateMessage(msg domain.Message) (*domain.Message, error)
	GetMessages() (*[]domain.Message, error)
}

func NewService(repo *repo.Repo) *Service {
	return &Service{
		Auth: newAuthService(repo.Auth),
		Chat: newChatService(repo.Chat),
	}
}

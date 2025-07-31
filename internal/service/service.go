package service

import (
	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/RX90/Chat/internal/repo"
	"github.com/google/uuid"
)

type Service struct {
	Auth AuthService
	Chat ChatService
}

type AuthService interface {
	CreateUser(user *entities.User) error
	GetUserByEmail(email string) (*entities.User, error)
	NewRefreshToken(userID uuid.UUID) (*entities.RefreshToken, error)
	CheckRefreshToken(userID uuid.UUID, refreshToken string) error
	DeleteRefreshToken(userID uuid.UUID) error
}

type ChatService interface {
	CreateMessage(msg *entities.Message) (*dto.CreatedMessage, error)
	GetMessages() (*[]dto.CreatedMessage, error)
}

func NewService(repo *repo.Repo) *Service {
	return &Service{
		Auth: newAuthService(repo.Auth),
		Chat: newChatService(repo.Chat),
	}
}

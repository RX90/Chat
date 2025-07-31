package repo

import (
	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	Auth AuthRepo
	Chat ChatRepo
}

type AuthRepo interface {
	CreateUser(user *entities.User) error
	GetUserByEmail(email string) (*entities.User, error)
	UpsertRefreshToken(token *entities.RefreshToken) error
	CheckRefreshToken(userID uuid.UUID, refreshToken string) error
	DeleteRefreshToken(userID uuid.UUID) error
}

type ChatRepo interface {
	CreateMessage(msg *entities.Message) (*dto.CreatedMessage, error)
	GetMessages() (*[]dto.CreatedMessage, error)
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		Auth: newAuthRepo(db),
		Chat: newChatRepo(db),
	}
}

package repo

import (
	"github.com/RX90/Chat/internal/domain"
	"gorm.io/gorm"
)

type Repo struct {
	Auth AuthRepo
	Chat ChatRepo
}

type AuthRepo interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	UpsertRefreshToken(token *domain.Token) error
}

type ChatRepo interface {
	CreateMessage(msg domain.Message) (*domain.Message, error)
	GetMessages() (*[]domain.Message, error)
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		Auth: newAuthRepo(db),
		Chat: newChatRepo(db),
	}
}

package repo

import (
	"github.com/RX90/Chat/internal/domain"
	"gorm.io/gorm"
)

type chatRepo struct {
	db *gorm.DB
}

func newChatRepo(db *gorm.DB) ChatRepo {
	return &chatRepo{db: db}
}

func (r *chatRepo) CreateMessage(msg domain.Message) (*domain.Message, error) {
	if err := r.db.Create(&msg).Error; err != nil {
		return &msg, err
	}
	return &msg, nil
}

func (r *chatRepo) GetMessages() (*[]domain.Message, error) {
	var msgs []domain.Message
	if err := r.db.Order("created_at ASC").Find(&msgs).Error; err != nil {
		return nil, err
	}
	return &msgs, nil
}
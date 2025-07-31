package repo

import (
	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"gorm.io/gorm"
)

type chatRepo struct {
	db *gorm.DB
}

func newChatRepo(db *gorm.DB) ChatRepo {
	return &chatRepo{db: db}
}

func (r *chatRepo) CreateMessage(msg *entities.Message) (dto.OutgoingMessage, error) {
	if err := r.db.Create(msg).Error; err != nil {
		return dto.OutgoingMessage{}, err
	}

	var msgOut dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.created_at, users.username").
		Joins("left join users on users.id = messages.user_id").
		Where("messages.id = ?", msg.ID).
		Scan(&msgOut).Error

	return msgOut, err
}

func (r *chatRepo) GetMessages() ([]dto.OutgoingMessage, error) {
	var msgs []dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.created_at, users.username").
		Joins("left join users on users.id = messages.user_id").
		Order("messages.created_at ASC").
		Scan(&msgs).Error

	return msgs, err
}

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

func (r *chatRepo) FindMessageByID(msgID int) (*entities.Message, error) {
	var msg entities.Message
	if err := r.db.First(&msg, msgID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepo) CreateMessage(msg *entities.Message) (dto.OutgoingMessage, error) {
	if err := r.db.Create(msg).Error; err != nil {
		return dto.OutgoingMessage{}, err
	}

	var msgOut dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.user_id, users.username, messages.created_at, messages.updated_at").
		Joins("left join users on users.id = messages.user_id").
		Where("messages.id = ?", msg.ID).
		Scan(&msgOut).Error

	return msgOut, err
}

func (r *chatRepo) GetMessages() ([]dto.OutgoingMessage, error) {
	var msgs []dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.user_id, users.username, messages.created_at, messages.updated_at").
		Joins("left join users on users.id = messages.user_id").
		Order("messages.created_at ASC").
		Scan(&msgs).Error

	return msgs, err
}

func (r *chatRepo) UpdateMessageContent(msg *entities.Message) (dto.OutgoingMessage, error) {
	if err := r.db.Save(&msg).Error; err != nil {
		return dto.OutgoingMessage{}, err
	}

	var msgOut dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.user_id, users.username, messages.created_at, messages.updated_at").
		Joins("left join users on users.id = messages.user_id").
		Where("messages.id = ?", msg.ID).
		Scan(&msgOut).Error
	if err != nil {
		return dto.OutgoingMessage{}, err
	}

	return msgOut, nil
}

func (r *chatRepo) DeleteMessageByID(msgID int) error {
	return r.db.Delete(&entities.Message{}, msgID).Error
}
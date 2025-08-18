package repo

import (
	"errors"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/google/uuid"
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
		Select("messages.id, messages.content, messages.user_id, users.username, messages.created_at").
		Joins("left join users on users.id = messages.user_id").
		Where("messages.id = ?", msg.ID).
		Scan(&msgOut).Error

	return msgOut, err
}

func (r *chatRepo) GetMessages() ([]dto.OutgoingMessage, error) {
	var msgs []dto.OutgoingMessage

	err := r.db.
		Table("messages").
		Select("messages.id, messages.content, messages.user_id, users.username, messages.created_at").
		Joins("left join users on users.id = messages.user_id").
		Order("messages.created_at ASC").
		Scan(&msgs).Error

	return msgs, err
}

func (r *chatRepo) DeleteMessage(msgID int, userID uuid.UUID) error {
	var msg entities.Message

	if err := r.db.First(&msg, msgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return err
	}

	if msg.UserID != userID {
		return errors.New("cannot delete another user's message")
	}

	if err := r.db.Delete(&msg).Error; err != nil {
		return err
	}

	return nil
}

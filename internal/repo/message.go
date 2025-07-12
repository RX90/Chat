package repo

import (
	"github.com/RX90/Chat/internal/domain"
)

func (r *Repo) CreateMessage(msg domain.Message) (domain.Message, error) {
	if err := r.db.Create(&msg).Error; err != nil {
		return msg, err
	}
	return msg, nil
}

func (r *Repo) GetMessages() ([]domain.Message, error) {
	var msgs []domain.Message
	if err := r.db.Order("created_at ASC").Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

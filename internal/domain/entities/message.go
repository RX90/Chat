package entities

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UserID    uuid.UUID `json:"user_id"`
}

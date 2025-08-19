package entities

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;autoCreateTime"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;autoUpdateTime"`
	UserID    uuid.UUID `gorm:"type:uuid;index;constraint:OnDelete:SET NULL;"`
}

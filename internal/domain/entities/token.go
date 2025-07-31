package entities

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	RefreshToken string    `gorm:"type:text;not null"`
	ExpiresAt    time.Time `gorm:"column:expires_at;not null"`
	User         User      `gorm:"constraint:OnDelete:CASCADE"`
}

package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"` 
	Username  string    `gorm:"type:text;not null"`
	Password  string    `gorm:"column:password_hash;type:text;not null"`
	Email     string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone;default:now()"`
}

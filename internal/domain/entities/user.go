package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `gorm:"column:password_hash" json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

package entities

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID           int
	UserID       uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
}
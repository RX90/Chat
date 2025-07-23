package domain

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           int
	UserID       uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
}
package domain

import "time"

type Message struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

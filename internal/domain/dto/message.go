package dto

import "time"

type CreatedMessage struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username"`
}

type WSClientMessage struct {
	Type    string `json:"type"`
	Token   string `json:"token,omitempty"`
	Content string `json:"content,omitempty"`
}

type WSServerMessage struct {
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	From      string    `json:"from,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

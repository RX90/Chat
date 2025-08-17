package dto

import "time"

type IncomingMessage struct {
	Type    string `json:"type"`
	Token   string `json:"token,omitempty"`
	Content string `json:"content,omitempty"`
}

type OutgoingMessage struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username"`
}

type OnlineUsersMessage struct {
    Type  string   `json:"type"`
    Users []string `json:"users"`
}
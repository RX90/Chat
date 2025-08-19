package dto

import "time"

type IncomingMessage struct {
	Type      string `json:"type"`
	Token     string `json:"token,omitempty"`
	Content   string `json:"content,omitempty"`
	MessageID int    `json:"messageId,omitempty"`
}

type OutgoingMessage struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateMessage struct {
	Type    string `json:"type"`
	Message OutgoingMessage
}

type DeleteMessage struct {
	Type      string `json:"type"`
	MessageID int    `json:"messageId"`
}

type OnlineUsersMessage struct {
	Type  string   `json:"type"`
	Users []string `json:"users"`
}

type AuthOK struct {
	Type string `json:"type"`
}

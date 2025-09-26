package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/RX90/Chat/internal/middleware"
	"github.com/RX90/Chat/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pingPeriod     = 20 * time.Second
	pongWait       = 30 * time.Second
	maxMessageSize = 4096
)

const (
	StateUnauthenticated = iota
	StateAuthenticated
	StateReady
)

var (
	newline       = []byte{'\n'}
	sendBufferCap = 256
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	send            chan []byte
	service         service.ChatService
	userID          uuid.UUID
	username        string
	tokenExpiryUnix int64
}

func ServeClient(conn *websocket.Conn, service service.ChatService) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	hub := getHub()

	client := &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, sendBufferCap),
		service: service,
	}

	hub.registerClient(client)

	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregisterClient(c)
		c.conn.Close()
	}()

	state := StateUnauthenticated

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure) {
				log.Printf("ws: unexpected close: %v", err)
			}
			break
		}

		var incoming dto.IncomingMessage
		if err := json.Unmarshal(msgBytes, &incoming); err != nil {
			log.Printf("invalid message format: %v", err)
			continue
		}

		log.Printf("WS incoming: %v", incoming.Type)

		switch state {
		case StateUnauthenticated:
			if incoming.Type != "auth" {
				c.closeWithPolicy("auth required")
				return
			}

			claims, err := middleware.ParseAccessToken(incoming.Token)
			if err != nil {
				c.closeWithPolicy("invalid token")
				return
			}

			c.userID = uuid.MustParse(claims.Subject)
			c.username = claims.Username
			c.setExpiry(time.Unix(claims.ExpiresAt, 0))

			go c.writePump()

			authMsg, _ := json.Marshal(dto.AuthOK{Type: "auth_ok"})
			c.sendMessage(authMsg)

			c.hub.broadcastOnlineUsers()

			state = StateAuthenticated

		case StateAuthenticated:
			if incoming.Type != "history" {
				c.closeWithPolicy("history required")
				return
			}

			if time.Now().After(c.getExpiry()) {
				c.closeWithPolicy("token has expired")
				return
			}

			history, err := c.service.GetMessages()
			if err != nil {
				log.Printf("failed to get history: %v", err)
				continue
			}

			for _, msg := range history {
				jsonMsg, err := json.Marshal(msg)
				if err != nil {
					log.Printf("msg marshal error: %v", err)
					continue
				}
				if err := c.sendMessage(jsonMsg); err != nil {
					log.Printf("send message error: %v", err)
					break
				}
			}

			state = StateReady

		case StateReady:
			switch incoming.Type {
			case "auth_refresh":
				claims, err := middleware.ParseAccessToken(incoming.Token)
				if err != nil {
					c.closeWithPolicy("invalid token")
					return
				}
				c.setExpiry(time.Unix(claims.ExpiresAt, 0))

			case "message":
				if time.Now().After(c.getExpiry()) {
					c.closeWithPolicy("token has expired")
					return
				}

				msg := entities.Message{
					Content: incoming.Content,
					UserID:  c.userID,
				}

				msgLen := utf8.RuneCountInString(incoming.Content)
				if msgLen == 0 || msgLen > 1024 {
					log.Printf("incorrect message length: %d", msgLen)
					continue
				}

				createdMsg, err := c.service.CreateMessage(&msg)
				if err != nil {
					log.Printf("failed to create message: %v", err)
					continue
				}

				jsonMsg, err := json.Marshal(createdMsg)
				if err != nil {
					log.Printf("marshal error: %v", err)
					continue
				}

				c.hub.broadcastMessage(jsonMsg)

			case "update":
				if time.Now().After(c.getExpiry()) {
					c.closeWithPolicy("token has expired")
					return
				}

				msgID := incoming.MessageID
				if msgID == 0 {
					c.closeWithPolicy("invalid message id")
					return
				}

				content := incoming.Content
				if content == "" {
					c.closeWithPolicy("empty content")
					return
				}

				updatedMsg, err := c.service.UpdateMessage(msgID, c.userID, content)
				if err != nil {
					log.Printf("failed to update message %v: %v", msgID, err)
					continue
				}

				outgoing := dto.UpdateMessage{
					Type:    "update",
					Message: updatedMsg,
				}

				jsonMsg, err := json.Marshal(outgoing)
				if err != nil {
					log.Printf("marshal error: %v", err)
					continue
				}

				c.hub.broadcastMessage(jsonMsg)

			case "delete":
				if time.Now().After(c.getExpiry()) {
					c.closeWithPolicy("token has expired")
					return
				}

				msgID := incoming.MessageID
				if msgID == 0 {
					c.closeWithPolicy("invalid message id")
					return
				}

				if err := c.service.DeleteMessage(msgID, c.userID); err != nil {
					log.Printf("failed to delete message %v: %v", msgID, err)
					continue
				}

				outgoing := dto.DeleteMessage{
					Type:      "delete",
					MessageID: msgID,
				}

				jsonMsg, err := json.Marshal(outgoing)
				if err != nil {
					log.Printf("marshal error: %v", err)
					continue
				}

				c.hub.broadcastMessage(jsonMsg)

			default:
				c.closeWithPolicy(fmt.Sprintf("unknown message type: %s", incoming.Type))
				return
			}

		default:
			log.Printf("unknown state: %s", incoming.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if time.Now().After(c.getExpiry()) {
				c.closeWithPolicy("token has expired")
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			for range len(c.send) {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMessage(msg []byte) error {
    select {
    case c.send <- msg:
        return nil
    case <-time.After(2 * time.Second):
        log.Printf("Client %v: send buffer full after timeout", c)
        c.closeWithPolicy("send buffer full")
        c.hub.unregisterClient(c)
        return errors.New("send buffer full")
    }
}

func (c *Client) setExpiry(t time.Time) {
	atomic.StoreInt64(&c.tokenExpiryUnix, t.Unix())
}

func (c *Client) getExpiry() time.Time {
	sec := atomic.LoadInt64(&c.tokenExpiryUnix)
	return time.Unix(sec, 0)
}

func (c *Client) closeWithPolicy(reason string) {
	c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, reason),
	)
}

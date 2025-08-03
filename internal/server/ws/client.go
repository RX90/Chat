package ws

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/RX90/Chat/internal/domain/dto"
	"github.com/RX90/Chat/internal/domain/entities"
	"github.com/RX90/Chat/internal/middleware"
	"github.com/RX90/Chat/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var (
	newline    = []byte{'\n'}
	bufferSize = 256
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:3000"
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	service     service.ChatService
	userID      uuid.UUID
	tokenExpiry time.Time
}

func NewClient(conn *websocket.Conn, service service.ChatService) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	_, authMsg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("failed to read auth message: %v", err)
		conn.Close()
		return
	}

	var init map[string]string
	if err := json.Unmarshal(authMsg, &init); err != nil {
		log.Printf("invalid auth message format: %v", err)
		conn.Close()
		return
	}

	if init["type"] != "auth" || init["token"] == "" {
		log.Println("missing auth token")
		conn.Close()
		return
	}

	claims, err := middleware.ParseAccessToken(init["token"])
	if err != nil {
		log.Printf("client auth failed: %v", err)
		conn.Close()
		return
	}

	h := getHub()

	c := &Client{
		hub:         h,
		conn:        conn,
		send:        make(chan []byte, bufferSize),
		service:     service,
		userID:      uuid.MustParse(claims.Subject),
		tokenExpiry: time.Unix(claims.ExpiresAt, 0),
	}

	h.registerClient(c)

	msg := []byte(`{"type":"auth_ok"}`)
	if err := c.SendMessage(msg); err != nil {
		log.Printf("send message error: %v", err)
		conn.Close()
		return
	}

	go c.readPump()
	go c.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregisterClient(c)
		c.conn.Close()
	}()

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				log.Printf("ws: unexpected close: %v", err)
			}
			break
		}

		var incoming dto.IncomingMessage
		if err := json.Unmarshal(msgBytes, &incoming); err != nil {
			log.Printf("invalid message format: %v", err)
			continue
		}

		switch incoming.Type {
		case "auth":
			claims, err := middleware.ParseAccessToken(incoming.Token)
			if err != nil {
				log.Printf("auth failed: %v", err)
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid token"))
				return
			}
			c.userID = uuid.MustParse(claims.Subject)
			c.tokenExpiry = time.Unix(claims.ExpiresAt, 0)
		case "history":
			if time.Now().After(c.tokenExpiry) {
				log.Printf("token expired for user %s", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "token expired"))
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
					log.Printf("marshal error: %v", err)
					continue
				}
				if err := c.SendMessage(jsonMsg); err != nil {
					log.Printf("send error: %v", err)
					break
				}
			}
		case "message":
			if time.Now().After(c.tokenExpiry) {
				log.Printf("token expired for user %s", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "token expired"))
				return
			}

			msg := entities.Message{
				Content: incoming.Content,
				UserID:  c.userID,
			}

			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Printf("message unmarshal error: %v", err)
				continue
			}

			createdMsg, err := c.service.CreateMessage(&msg)
			if err != nil {
				log.Printf("failed to create message: %v", err)
				break
			}

			jsonMsg, err := json.Marshal(createdMsg)
			if err != nil {
				log.Printf("marshal error: %v", err)
				continue
			}

			c.hub.broadcastMessage(jsonMsg)
		default:
			log.Printf("unknown message type: %s", incoming.Type)
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

			if time.Now().After(c.tokenExpiry) {
				log.Printf("token expired during send for user %s", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "token expired"))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for range n {
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

func (c *Client) SendMessage(msg []byte) error {
	select {
	case c.send <- msg:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

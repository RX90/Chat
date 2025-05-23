package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

type Chat struct {
	Server *Server
	Hub    *Hub
	DB     *sqlx.DB
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	db     *sqlx.DB
	userID string
}

type Message struct {
	Content     string `json:"content" db:"content"`
	SenderLogin string `json:"sender" db:"sender"`
	Time        string `json:"time" db:"created_at"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	newline = []byte{'\n'}
)

func (c *Chat) handleChatPage(ctx *gin.Context) {
	http.ServeFile(ctx.Writer, ctx.Request, "chat.html")
}

func (c *Chat) handleChatWS(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("can't upgrade http request: %v", err)})
		return
	}
	client := &Client{hub: c.Hub, conn: conn, send: make(chan []byte, 256), db: c.DB, userID: userID}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	messages, err := c.loadChatHistory()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to load chat history: %v", err)})
		return
	} else {
		for _, msg := range messages {
			outgoingMsg := Message{
				Content:     msg.Content,
				SenderLogin: msg.SenderLogin,
				Time:        msg.Time,
			}
			jsonMsg, err := json.Marshal(outgoingMsg)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("failed to marshal history message: %v", err)})
				return
			}
			client.send <- jsonMsg
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

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incomingMsg Message
		if err := json.Unmarshal(msgBytes, &incomingMsg); err != nil {
			log.Printf("invalid json: %v", err)
			continue
		}

		_, err = c.db.Exec(
			"INSERT INTO messages (sender_id, created_at, content) VALUES ($1, $2, $3)",
			c.userID,
			incomingMsg.Time,
			strings.TrimSpace(incomingMsg.Content),
		)
		if err != nil {
			log.Printf("Failed to save message to DB: %v", err)
			continue
		}

		outgoingMsg := Message{
			Content:     strings.TrimSpace(incomingMsg.Content),
			Time:        incomingMsg.Time,
		}

		jsonMsg, err := json.Marshal(outgoingMsg)
		if err != nil {
			log.Printf("Failed to marshal json: %v", err)
			continue
		}

		c.hub.broadcast <- jsonMsg
	}
}

func (c *Chat) loadChatHistory() ([]Message, error) {
	var messages []Message

	query := `
        SELECT m.content, u.login AS sender, m.created_at
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        ORDER BY m.created_at ASC
    `
	err := c.DB.Select(&messages, query)

	return messages, err
}

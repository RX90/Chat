package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	ctx.HTML(http.StatusOK, "chat.html", nil)
}

func (c *Chat) handleChatWS(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	log.Println(userID)
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
			jsonMsg, err := json.Marshal(msg)
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
				log.Printf("unexpected close in readPump: %v", err)
				continue
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("failed to unmarshall JSON in readPump: %v", err)
			continue
		}

		msg.Time = time.Now().UTC().Format(time.RFC3339)

		if err = c.db.QueryRowx(`
    		WITH inserted_message AS (
        		INSERT INTO messages (sender_id, created_at, content) 
        		VALUES ($1, $2, $3)
        		RETURNING id, content, created_at
    		)
    		SELECT im.content, u.login AS sender, to_char(im.created_at AT TIME ZONE 'Asia/Yekaterinburg', 'YYYY-MM-DD HH24:MI:SS') AS created_at
    		FROM inserted_message im
    		JOIN users u ON u.id = $1`,
			c.userID,
			msg.Time,
			msg.Content,
		).StructScan(&msg); err != nil {
			log.Printf("DB error while inserting message from readPump: %v", err)
			continue
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("failed to marshal JSON in readPump: %v", err)
			continue
		}

		c.hub.broadcast <- jsonMsg
	}
}

func (c *Chat) loadChatHistory() ([]Message, error) {
	var messages []Message

	query := `
        SELECT m.content, u.login AS sender, to_char(m.created_at AT TIME ZONE 'Asia/Yekaterinburg', 'YYYY-MM-DD HH24:MI:SS') AS created_at
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        ORDER BY m.created_at ASC
    `
	err := c.DB.Select(&messages, query)

	return messages, err
}

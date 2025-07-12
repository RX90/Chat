package ws

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/repo"
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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	c := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, bufferSize),
	}
	hub.RegisterClient(c)
	return c
}

func (c *Client) ReadPump(repo *repo.Repo) {
	defer func() {
		c.hub.UnregisterClient(c)
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
				log.Printf("ws: unexpected close: %v", err)
			}
			break
		}
		var msg domain.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("failed to unmarshall JSON in readPump: %v", err)
			continue
		}

		msg, err = repo.CreateMessage(msg)
		if err != nil {
			log.Printf("failed to create message: %v", err)
			break
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("failed to marshal JSON after CreateMessage: %v", err)
			continue
		}

		c.hub.BroadcastMessage(jsonMsg)
	}
}

func (c *Client) WritePump() {
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

func (c *Client) SendMessage(msg []byte) error {
	select {
	case c.send <- msg:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

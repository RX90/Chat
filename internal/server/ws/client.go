package ws

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/RX90/Chat/internal/domain"
	"github.com/RX90/Chat/internal/service"
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
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	service service.ChatService
}

func NewClient(conn *websocket.Conn, service service.ChatService) *Client {
	h := getHub()
	c := &Client{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, bufferSize),
		service: service,
	}
	h.registerClient(c)
	go c.readPump()
	go c.writePump()
	return c
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregisterClient(c)
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
		updatedMsg, err := c.service.CreateMessage(msg)
		if err != nil {
			log.Printf("failed to create message: %v", err)
			break
		}

		jsonMsg, err := json.Marshal(updatedMsg)
		if err != nil {
			log.Printf("failed to marshal JSON after CreateMessage: %v", err)
			continue
		}

		c.hub.broadcastMessage(jsonMsg)
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

func (c *Client) SendMessage(msg []byte) error {
	select {
	case c.send <- msg:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

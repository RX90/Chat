package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/RX90/Chat/internal/domain/dto"
)

var (
	hubInstance        *Hub
	hubOnce            sync.Once
	broadcastBufferCap = 256
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
}

func getHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			broadcast:  make(chan []byte, broadcastBufferCap),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			clients:    make(map[*Client]bool),
			quit:       make(chan struct{}),
		}
		go hubInstance.run()
	})
	return hubInstance
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

			h.broadcastOnlineUsers()

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case <-h.quit:
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			return
		}
	}
}

func (h *Hub) registerClient(c *Client) {
	h.register <- c
}

func (h *Hub) unregisterClient(c *Client) {
	h.unregister <- c
}

func (h *Hub) broadcastMessage(msg []byte) {
	h.broadcast <- msg
}

func (h *Hub) getOnlineUsernames() []string {
	unique := make(map[string]struct{})

	for client := range h.clients {
		if client.username != "" {
			unique[client.username] = struct{}{}
		}
	}

	users := make([]string, 0, len(unique))

	for username := range unique {
		users = append(users, username)
	}
	return users
}

func (h *Hub) broadcastOnlineUsers() {
	users := h.getOnlineUsernames()
	msg := dto.OnlineUsersMessage{
		Type:  "online_users",
		Users: users,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error for online users: %v", err)
		return
	}

	h.broadcastMessage(data)
}

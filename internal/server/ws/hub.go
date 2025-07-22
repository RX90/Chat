package ws

import "sync"

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
}

var (
	hubInstance *Hub
	hubOnce     sync.Once
)

func getHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			broadcast:  make(chan []byte),
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

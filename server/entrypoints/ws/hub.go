package ws

import (
	"sync"
)

type Client struct {
	ID     string
	UserID string
	Send   chan []byte
}

type Hub struct {
	clients    map[string]*Client
	userIndex  map[string][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
	mu         sync.RWMutex
}

type broadcastMsg struct {
	userID string
	data   []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		userIndex:  make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMsg),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.userIndex[client.UserID] = append(h.userIndex[client.UserID], client)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)

				// Remove from user index
				clients := h.userIndex[client.UserID]
				for i, c := range clients {
					if c.ID == client.ID {
						h.userIndex[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.userIndex[msg.userID]
			for _, client := range clients {
				select {
				case client.Send <- msg.data:
				default:
					// Buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToUser(userID string, data []byte) {
	h.broadcast <- broadcastMsg{userID: userID, data: data}
}

func (h *Hub) BroadcastToAll(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

package ws

import (
	"encoding/json"
	"sync"
)

type Client struct {
	ID      string
	UserID  string
	AgentID string
	Send    chan []byte
}

type Hub struct {
	clients    map[string]*Client
	userIndex  map[string][]*Client
	agentIndex map[string]*Client
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
		agentIndex: make(map[string]*Client),
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
			if client.AgentID != "" {
				h.agentIndex[client.AgentID] = client
			}
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

				// Remove from agent index
				if client.AgentID != "" {
					delete(h.agentIndex, client.AgentID)
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
	// Copy client list under lock to avoid holding lock during send
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Send to all clients without holding the lock
	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) BroadcastToAgent(agentID string, msgType string, payload interface{}) error {
	msg, err := NewMessage(MessageType(msgType), payload)
	if err != nil {
		return err
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.mu.RLock()
	client, ok := h.agentIndex[agentID]
	h.mu.RUnlock()

	if !ok {
		return nil // Agent not connected, message will be lost (or could queue)
	}

	select {
	case client.Send <- data:
	default:
		// Buffer full
	}

	return nil
}

func (h *Hub) SetAgentID(clientID, agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[clientID]; ok {
		// Remove old agent mapping if exists
		if client.AgentID != "" {
			delete(h.agentIndex, client.AgentID)
		}
		client.AgentID = agentID
		h.agentIndex[agentID] = client
	}
}

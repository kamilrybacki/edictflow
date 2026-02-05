package worker

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure for production
	},
}

// Handler handles WebSocket connections for workers
type Handler struct {
	hub *Hub
}

// NewHandler creates a new worker WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeHTTP upgrades to WebSocket and manages the connection
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get team ID from query or header
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		teamID = r.Header.Get("X-Team-ID")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	agent := &AgentConn{
		ID:     uuid.New().String(),
		UserID: userID,
		TeamID: teamID,
		Send:   make(chan []byte, 256),
		conn:   conn,
	}

	h.hub.Register(agent)

	go h.writePump(agent)
	go h.readPump(agent)
}

func (h *Handler) readPump(agent *AgentConn) {
	defer func() {
		h.hub.Unregister(agent)
		agent.conn.Close()
	}()

	agent.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	agent.conn.SetPongHandler(func(string) error {
		agent.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := agent.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.handleMessage(agent, data)
	}
}

func (h *Handler) handleMessage(agent *AgentConn, data []byte) {
	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Invalid message: %v", err)
		return
	}

	switch msg.Type {
	case "heartbeat":
		var payload struct {
			AgentID string `json:"agent_id"`
			TeamID  string `json:"team_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			// Update agent info if provided
			if payload.AgentID != "" && agent.AgentID == "" {
				agent.AgentID = payload.AgentID
			}
			if payload.TeamID != "" && agent.TeamID != payload.TeamID {
				// Re-register with new team
				h.hub.Unregister(agent)
				agent.TeamID = payload.TeamID
				agent.Send = make(chan []byte, 256)
				h.hub.Register(agent)
			}
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	// Send ack
	ack, _ := json.Marshal(map[string]string{"type": "ack"})
	select {
	case agent.Send <- ack:
	default:
		// Buffer full, skip ack
	}
}

func (h *Handler) writePump(agent *AgentConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		agent.conn.Close()
	}()

	for {
		select {
		case message, ok := <-agent.Send:
			agent.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				agent.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := agent.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			agent.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := agent.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

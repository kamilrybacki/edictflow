package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/redis/go-redis/v9"
)

// AgentConn represents a connected agent
type AgentConn struct {
	ID      string
	UserID  string
	AgentID string
	TeamID  string
	Send    chan []byte
	conn    *websocket.Conn
}

// Hub manages agent connections and Redis subscriptions
type Hub struct {
	redisClient *redisAdapter.Client

	// Team -> agents mapping
	teamAgents map[string]map[*AgentConn]struct{}

	// Agent ID -> connection
	agents map[string]*AgentConn

	// Active Redis subscriptions per team
	subscriptions map[string]*redis.PubSub

	// Channels for goroutine communication
	register   chan *AgentConn
	unregister chan *AgentConn

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub creates a new worker hub
func NewHub(redisClient *redisAdapter.Client) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		redisClient:   redisClient,
		teamAgents:    make(map[string]map[*AgentConn]struct{}),
		agents:        make(map[string]*AgentConn),
		subscriptions: make(map[string]*redis.PubSub),
		register:      make(chan *AgentConn),
		unregister:    make(chan *AgentConn),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			h.cleanup()
			return

		case agent := <-h.register:
			h.handleRegister(agent)

		case agent := <-h.unregister:
			h.handleUnregister(agent)
		}
	}
}

func (h *Hub) handleRegister(agent *AgentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to agents map
	if agent.AgentID != "" {
		h.agents[agent.AgentID] = agent
	}

	// Add to team mapping
	if agent.TeamID != "" {
		if h.teamAgents[agent.TeamID] == nil {
			h.teamAgents[agent.TeamID] = make(map[*AgentConn]struct{})
		}
		h.teamAgents[agent.TeamID][agent] = struct{}{}

		// Subscribe to team channel if first agent for this team
		if len(h.teamAgents[agent.TeamID]) == 1 {
			h.subscribeToTeam(agent.TeamID)
		}
	}

	log.Printf("Agent registered: id=%s team=%s (total: %d)", agent.AgentID, agent.TeamID, len(h.agents))
}

func (h *Hub) handleUnregister(agent *AgentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from agents map
	if agent.AgentID != "" {
		delete(h.agents, agent.AgentID)
	}

	// Remove from team mapping
	if agent.TeamID != "" {
		if agents, ok := h.teamAgents[agent.TeamID]; ok {
			delete(agents, agent)

			// Unsubscribe if no agents left for this team
			if len(agents) == 0 {
				h.unsubscribeFromTeam(agent.TeamID)
				delete(h.teamAgents, agent.TeamID)
			}
		}
	}

	// Safe close - only close if not already closed
	select {
	case <-agent.Send:
		// Channel already closed
	default:
		close(agent.Send)
	}
	log.Printf("Agent unregistered: id=%s (remaining: %d)", agent.AgentID, len(h.agents))
}

func (h *Hub) subscribeToTeam(teamID string) {
	channel := events.ChannelForTeam(teamID)
	sub := h.redisClient.Subscribe(h.ctx, channel)
	h.subscriptions[teamID] = sub

	go h.listenToSubscription(teamID, sub)
	log.Printf("Subscribed to team channel: %s", channel)
}

func (h *Hub) unsubscribeFromTeam(teamID string) {
	if sub, ok := h.subscriptions[teamID]; ok {
		sub.Close()
		delete(h.subscriptions, teamID)
		log.Printf("Unsubscribed from team channel: team:%s:rules", teamID)
	}
}

func (h *Hub) listenToSubscription(teamID string, sub *redis.PubSub) {
	ch := sub.Channel()
	for msg := range ch {
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		h.broadcastToTeam(teamID, event)
	}
}

func (h *Hub) broadcastToTeam(teamID string, event events.Event) {
	h.mu.RLock()
	agents := h.teamAgents[teamID]
	h.mu.RUnlock()

	// Convert event to WebSocket message format
	wsMsg := map[string]interface{}{
		"type":      "config_update",
		"event":     event.Type,
		"entity_id": event.EntityID,
		"version":   event.Version,
		"timestamp": event.Timestamp,
	}
	data, _ := json.Marshal(wsMsg)

	for agent := range agents {
		select {
		case agent.Send <- data:
		default:
			// Buffer full, skip
		}
	}

	log.Printf("Broadcast %s to %d agents in team %s", event.Type, len(agents), teamID)
}

// Register adds an agent to the hub
func (h *Hub) Register(agent *AgentConn) {
	h.register <- agent
}

// Unregister removes an agent from the hub
func (h *Hub) Unregister(agent *AgentConn) {
	h.unregister <- agent
}

// Stop gracefully shuts down the hub
func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all subscriptions
	for _, sub := range h.subscriptions {
		sub.Close()
	}

	// Close all agent connections
	for _, agent := range h.agents {
		select {
		case <-agent.Send:
			// Already closed
		default:
			close(agent.Send)
		}
	}
}

// Stats returns hub statistics
func (h *Hub) Stats() (agents int, teams int, subscriptions int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.agents), len(h.teamAgents), len(h.subscriptions)
}

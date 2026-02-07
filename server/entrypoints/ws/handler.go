package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096, // Larger buffer for better throughput
	WriteBufferSize: 4096, // Larger buffer for better throughput
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

type Handler struct {
	hub            *Hub
	messageHandler MessageHandler
}

type MessageHandler interface {
	HandleHeartbeat(client *Client, payload HeartbeatPayload) error
	HandleDriftReport(client *Client, payload DriftReportPayload) error
	HandleContextDetected(client *Client, payload ContextDetectedPayload) error
	HandleSyncComplete(client *Client, payload SyncCompletePayload) error
	HandleChangeDetected(client *Client, payload ChangeDetectedPayload) error
	HandleChangeUpdated(client *Client, payload ChangeUpdatedPayload) error
	HandleExceptionRequest(client *Client, payload ExceptionRequestPayload) error
	HandleRevertComplete(client *Client, payload RevertCompletePayload) error
}

func NewHandler(hub *Hub, messageHandler MessageHandler) *Handler {
	return &Handler{
		hub:            hub,
		messageHandler: messageHandler,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

func (h *Handler) readPump(conn *websocket.Conn, client *Client) {
	defer func() {
		h.hub.Unregister(client)
		conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		h.handleMessage(client, msg)
	}
}

func (h *Handler) handleMessage(client *Client, msg Message) {
	if h.messageHandler != nil {
		switch msg.Type {
		case TypeHeartbeat:
			var payload HeartbeatPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleHeartbeat(client, payload)
			}

		case TypeDriftReport:
			var payload DriftReportPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleDriftReport(client, payload)
			}

		case TypeContextDetected:
			var payload ContextDetectedPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleContextDetected(client, payload)
			}

		case TypeSyncComplete:
			var payload SyncCompletePayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleSyncComplete(client, payload)
			}

		case TypeChangeDetected:
			var payload ChangeDetectedPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleChangeDetected(client, payload)
			}

		case TypeChangeUpdated:
			var payload ChangeUpdatedPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleChangeUpdated(client, payload)
			}

		case TypeExceptionRequest:
			var payload ExceptionRequestPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleExceptionRequest(client, payload)
			}

		case TypeRevertComplete:
			var payload RevertCompletePayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				_ = h.messageHandler.HandleRevertComplete(client, payload)
			}
		}
	}

	// Send ack
	ack, _ := NewMessage(TypeAck, AckPayload{RefID: msg.ID})
	ackData, _ := json.Marshal(ack)
	client.Send <- ackData
}

func (h *Handler) writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

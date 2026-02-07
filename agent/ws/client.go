// agent/ws/client.go
package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ErrBufferFull is returned when the send buffer is full and cannot accept more messages
var ErrBufferFull = errors.New("send buffer full, message dropped")

type State int

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
)

type MessageHandler func(Message)

type Client struct {
	serverURL      string
	token          string
	conn           *websocket.Conn
	state          State
	stateMu        sync.RWMutex
	send           chan []byte
	done           chan struct{}
	handlers       map[MessageType]MessageHandler
	handlersLock   sync.RWMutex
	onConnect      func()
	onDisconnect   func()
	reconnectDelay time.Duration
	maxReconnect   time.Duration
}

func NewClient(serverURL, token string) *Client {
	wsURL := serverURL
	if len(wsURL) > 5 && wsURL[:5] == "http:" {
		wsURL = "ws:" + wsURL[5:]
	} else if len(wsURL) > 6 && wsURL[:6] == "https:" {
		wsURL = "wss:" + wsURL[6:]
	}

	return &Client{
		serverURL:      wsURL + "/ws",
		token:          token,
		state:          StateDisconnected,
		send:           make(chan []byte, 256),
		handlers:       make(map[MessageType]MessageHandler),
		reconnectDelay: time.Second,
		maxReconnect:   60 * time.Second,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) OnConnect(fn func()) {
	c.onConnect = fn
}

func (c *Client) OnDisconnect(fn func()) {
	c.onDisconnect = fn
}

func (c *Client) OnMessage(msgType MessageType, handler MessageHandler) {
	c.handlersLock.Lock()
	c.handlers[msgType] = handler
	c.handlersLock.Unlock()
}

func (c *Client) State() State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

func (c *Client) Connect() error {
	c.stateMu.Lock()
	c.state = StateConnecting
	c.stateMu.Unlock()

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, header)
	if err != nil {
		c.stateMu.Lock()
		c.state = StateDisconnected
		c.stateMu.Unlock()
		return err
	}

	c.conn = conn
	c.done = make(chan struct{})
	c.reconnectDelay = time.Second

	// Set up pong handler to refresh read deadline on pong received
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	c.stateMu.Lock()
	c.state = StateConnected
	c.stateMu.Unlock()

	if c.onConnect != nil {
		c.onConnect()
	}

	go c.readPump()
	go c.writePump()

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.stateMu.Lock()
	c.state = StateDisconnected
	c.stateMu.Unlock()
}

func (c *Client) Send(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	default:
		return ErrBufferFull
	}
}

func (c *Client) readPump() {
	defer func() {
		c.stateMu.Lock()
		c.state = StateDisconnected
		c.stateMu.Unlock()
		close(c.done)
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}()

	// Set initial read deadline - will be refreshed by pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		// Refresh read deadline on any successful read
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		c.handlersLock.RLock()
		handler, ok := c.handlers[msg.Type]
		c.handlersLock.RUnlock()

		if ok {
			handler(msg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case data := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

// ConnectWithRetry attempts to maintain a persistent connection with exponential backoff.
// Deprecated: Use ConnectWithContext for proper shutdown handling.
func (c *Client) ConnectWithRetry() {
	c.ConnectWithContext(context.Background())
}

// ConnectWithContext attempts to maintain a persistent connection with exponential backoff.
// The context can be cancelled to gracefully stop the reconnection loop.
func (c *Client) ConnectWithContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket connection loop stopped by context cancellation")
			return
		default:
		}

		if err := c.Connect(); err != nil {
			log.Printf("Connection failed: %v, retrying in %v", err, c.reconnectDelay)
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.reconnectDelay):
			}
			c.reconnectDelay *= 2
			if c.reconnectDelay > c.maxReconnect {
				c.reconnectDelay = c.maxReconnect
			}
			continue
		}

		// Reset delay after successful connection
		c.reconnectDelay = time.Second

		// Wait for disconnection or context cancellation
		select {
		case <-ctx.Done():
			c.Close()
			return
		case <-c.done:
			log.Println("Connection closed, reconnecting...")
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.reconnectDelay):
			}
		}
	}
}

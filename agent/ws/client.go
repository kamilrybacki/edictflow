// agent/ws/client.go
package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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
		return nil // Drop if buffer full
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

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

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

func (c *Client) ConnectWithRetry() {
	for {
		if err := c.Connect(); err != nil {
			log.Printf("Connection failed: %v, retrying in %v", err, c.reconnectDelay)
			time.Sleep(c.reconnectDelay)
			c.reconnectDelay *= 2
			if c.reconnectDelay > c.maxReconnect {
				c.reconnectDelay = c.maxReconnect
			}
			continue
		}
		<-c.done
		log.Println("Connection closed, reconnecting...")
		time.Sleep(c.reconnectDelay)
	}
}

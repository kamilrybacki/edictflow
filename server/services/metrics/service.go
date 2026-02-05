package metrics

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/kamilrybacki/claudeception/server/adapters/splunk"
)

// Service defines the metrics service interface
type Service interface {
	// API metrics
	RecordAPIRequest(method, path string, statusCode int, duration time.Duration, userID string)
	RecordAPIError(method, path string, errorType string, userID string)

	// Redis pub/sub metrics
	RecordRedisEvent(eventType string, channel string, success bool)
	RecordRedisPublish(channel string, messageType string, success bool, latencyMs int64)
	RecordRedisSubscription(channel string, action string)

	// WebSocket hub metrics
	RecordHubStats(agents, teams, subscriptions int)
	RecordAgentConnection(agentID, teamID string, action string)
	RecordWebSocketMessage(direction string, messageType string, agentID string, sizeBytes int)
	RecordBroadcast(teamID string, eventType string, recipientCount int)

	// Database metrics
	RecordDBQuery(operation string, table string, duration time.Duration, success bool)
	RecordDBPoolStats(totalConns, acquiredConns, idleConns, maxConns int32)

	// System health metrics
	RecordHealthCheck(component string, status string, latencyMs int64)
	RecordWorkerHeartbeat(workerID string, agentCount, teamCount int)

	Flush() error
	Close() error
}

// Config holds metrics service configuration
type Config struct {
	SplunkConfig  splunk.Config
	BufferSize    int
	FlushInterval time.Duration
	Hostname      string
}

// SplunkService implements Service with Splunk HEC backend
type SplunkService struct {
	client        *splunk.Client
	hostname      string
	buffer        []splunk.Event
	bufferSize    int
	flushInterval time.Duration
	mu            sync.Mutex
	done          chan struct{}
	wg            sync.WaitGroup
}

// NewSplunkService creates a new Splunk-backed metrics service
func NewSplunkService(cfg Config) *SplunkService {
	bufferSize := cfg.BufferSize
	if bufferSize == 0 {
		bufferSize = 100
	}

	flushInterval := cfg.FlushInterval
	if flushInterval == 0 {
		flushInterval = 10 * time.Second
	}

	s := &SplunkService{
		client:        splunk.NewClient(cfg.SplunkConfig),
		hostname:      cfg.Hostname,
		buffer:        make([]splunk.Event, 0, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
	}

	s.wg.Add(1)
	go s.backgroundFlusher()

	return s
}

func (s *SplunkService) backgroundFlusher() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Flush(); err != nil {
				log.Printf("metrics flush error: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

// RecordAPIRequest records an HTTP API request metric
func (s *SplunkService) RecordAPIRequest(method, path string, statusCode int, duration time.Duration, userID string) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":        "api_request",
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"user_id":     userID,
		},
	}
	s.addEvent(event)
}

// RecordAPIError records an API error metric
func (s *SplunkService) RecordAPIError(method, path string, errorType string, userID string) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":       "api_error",
			"method":     method,
			"path":       path,
			"error_type": errorType,
			"user_id":    userID,
		},
	}
	s.addEvent(event)
}

// RecordRedisEvent records a Redis pub/sub event metric
func (s *SplunkService) RecordRedisEvent(eventType string, channel string, success bool) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":       "redis_event",
			"event_type": eventType,
			"channel":    channel,
			"success":    success,
		},
	}
	s.addEvent(event)
}

// RecordHubStats records WebSocket hub statistics
func (s *SplunkService) RecordHubStats(agents, teams, subscriptions int) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":          "hub_stats",
			"agents":        agents,
			"teams":         teams,
			"subscriptions": subscriptions,
		},
	}
	s.addEvent(event)
}

// RecordDBQuery records a database query metric
func (s *SplunkService) RecordDBQuery(operation string, table string, duration time.Duration, success bool) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":        "db_query",
			"operation":   operation,
			"table":       table,
			"duration_ms": duration.Milliseconds(),
			"success":     success,
		},
	}
	s.addEvent(event)
}

// RecordRedisPublish records a Redis publish operation with latency
func (s *SplunkService) RecordRedisPublish(channel string, messageType string, success bool, latencyMs int64) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":         "redis_publish",
			"channel":      channel,
			"message_type": messageType,
			"success":      success,
			"latency_ms":   latencyMs,
		},
	}
	s.addEvent(event)
}

// RecordRedisSubscription records Redis subscription changes
func (s *SplunkService) RecordRedisSubscription(channel string, action string) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":    "redis_subscription",
			"channel": channel,
			"action":  action,
		},
	}
	s.addEvent(event)
}

// RecordAgentConnection records agent connect/disconnect events
func (s *SplunkService) RecordAgentConnection(agentID, teamID string, action string) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":     "agent_connection",
			"agent_id": agentID,
			"team_id":  teamID,
			"action":   action,
		},
	}
	s.addEvent(event)
}

// RecordWebSocketMessage records WebSocket message metrics
func (s *SplunkService) RecordWebSocketMessage(direction string, messageType string, agentID string, sizeBytes int) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":         "websocket_message",
			"direction":    direction,
			"message_type": messageType,
			"agent_id":     agentID,
			"size_bytes":   sizeBytes,
		},
	}
	s.addEvent(event)
}

// RecordBroadcast records broadcast events to teams
func (s *SplunkService) RecordBroadcast(teamID string, eventType string, recipientCount int) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":            "broadcast",
			"team_id":         teamID,
			"event_type":      eventType,
			"recipient_count": recipientCount,
		},
	}
	s.addEvent(event)
}

// RecordDBPoolStats records database connection pool statistics
func (s *SplunkService) RecordDBPoolStats(totalConns, acquiredConns, idleConns, maxConns int32) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":           "db_pool_stats",
			"total_conns":    totalConns,
			"acquired_conns": acquiredConns,
			"idle_conns":     idleConns,
			"max_conns":      maxConns,
		},
	}
	s.addEvent(event)
}

// RecordHealthCheck records health check results for components
func (s *SplunkService) RecordHealthCheck(component string, status string, latencyMs int64) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":       "health_check",
			"component":  component,
			"status":     status,
			"latency_ms": latencyMs,
		},
	}
	s.addEvent(event)
}

// RecordWorkerHeartbeat records worker heartbeat with current stats
func (s *SplunkService) RecordWorkerHeartbeat(workerID string, agentCount, teamCount int) {
	event := splunk.Event{
		Host: s.hostname,
		Event: map[string]interface{}{
			"type":        "worker_heartbeat",
			"worker_id":   workerID,
			"agent_count": agentCount,
			"team_count":  teamCount,
		},
	}
	s.addEvent(event)
}

func (s *SplunkService) addEvent(event splunk.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer = append(s.buffer, event)

	if len(s.buffer) >= s.bufferSize {
		go func() {
			if err := s.Flush(); err != nil {
				log.Printf("metrics flush error: %v", err)
			}
		}()
	}
}

// Flush sends all buffered events to Splunk
func (s *SplunkService) Flush() error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}

	events := make([]splunk.Event, len(s.buffer))
	copy(events, s.buffer)
	s.buffer = s.buffer[:0]
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.client.SendBatch(ctx, events)
}

// Close flushes remaining events and stops the background flusher
func (s *SplunkService) Close() error {
	close(s.done)
	s.wg.Wait()
	return s.Flush()
}

// NoOpService is a no-operation metrics service for when metrics are disabled
type NoOpService struct{}

func (n *NoOpService) RecordAPIRequest(method, path string, statusCode int, duration time.Duration, userID string) {
}
func (n *NoOpService) RecordAPIError(method, path string, errorType string, userID string)      {}
func (n *NoOpService) RecordRedisEvent(eventType string, channel string, success bool)         {}
func (n *NoOpService) RecordRedisPublish(channel, messageType string, success bool, latencyMs int64) {
}
func (n *NoOpService) RecordRedisSubscription(channel string, action string)                          {}
func (n *NoOpService) RecordHubStats(agents, teams, subscriptions int)                                {}
func (n *NoOpService) RecordAgentConnection(agentID, teamID string, action string)                    {}
func (n *NoOpService) RecordWebSocketMessage(direction, messageType, agentID string, sizeBytes int)   {}
func (n *NoOpService) RecordBroadcast(teamID string, eventType string, recipientCount int)            {}
func (n *NoOpService) RecordDBQuery(operation, table string, duration time.Duration, success bool)    {}
func (n *NoOpService) RecordDBPoolStats(totalConns, acquiredConns, idleConns, maxConns int32)         {}
func (n *NoOpService) RecordHealthCheck(component string, status string, latencyMs int64)             {}
func (n *NoOpService) RecordWorkerHeartbeat(workerID string, agentCount, teamCount int)               {}
func (n *NoOpService) Flush() error                                                                   { return nil }
func (n *NoOpService) Close() error                                                                   { return nil }

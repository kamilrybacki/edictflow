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
	RecordAPIRequest(method, path string, statusCode int, duration time.Duration, userID string)
	RecordAPIError(method, path string, errorType string, userID string)
	RecordRedisEvent(eventType string, channel string, success bool)
	RecordHubStats(agents, teams, subscriptions int)
	RecordDBQuery(operation string, table string, duration time.Duration, success bool)
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
func (n *NoOpService) RecordAPIError(method, path string, errorType string, userID string) {}
func (n *NoOpService) RecordRedisEvent(eventType string, channel string, success bool)    {}
func (n *NoOpService) RecordHubStats(agents, teams, subscriptions int)                    {}
func (n *NoOpService) RecordDBQuery(operation string, table string, duration time.Duration, success bool) {
}
func (n *NoOpService) Flush() error { return nil }
func (n *NoOpService) Close() error { return nil }

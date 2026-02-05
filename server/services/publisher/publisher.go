package publisher

import (
	"context"
	"log"
	"time"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/metrics"
)

// Publisher broadcasts events to Redis channels
type Publisher interface {
	PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error
	PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error
	PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error
	PublishToAgent(ctx context.Context, agentID string, data []byte) error
}

// RedisPublisher implements Publisher using Redis
type RedisPublisher struct {
	client  *redisAdapter.Client
	metrics metrics.Service
}

// NewRedisPublisher creates a new publisher
func NewRedisPublisher(client *redisAdapter.Client) *RedisPublisher {
	return &RedisPublisher{
		client:  client,
		metrics: &metrics.NoOpService{},
	}
}

// SetMetrics sets the metrics service for observability
func (p *RedisPublisher) SetMetrics(m metrics.Service) {
	p.metrics = m
}

// PublishRuleEvent publishes a rule change event
func (p *RedisPublisher) PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error {
	event := events.NewEvent(eventType, ruleID, teamID)
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	channel := events.ChannelForTeam(teamID)
	log.Printf("Publishing %s to %s: rule=%s", eventType, channel, ruleID)

	start := time.Now()
	err = p.client.Publish(ctx, channel, data)
	latency := time.Since(start).Milliseconds()
	p.metrics.RecordRedisPublish(channel, string(eventType), err == nil, latency)

	return err
}

// PublishCategoryEvent publishes a category change event
func (p *RedisPublisher) PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error {
	event := events.NewEvent(eventType, categoryID, teamID)
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	channel := events.ChannelForTeamCategories(teamID)
	log.Printf("Publishing %s to %s: category=%s", eventType, channel, categoryID)

	start := time.Now()
	err = p.client.Publish(ctx, channel, data)
	latency := time.Since(start).Milliseconds()
	p.metrics.RecordRedisPublish(channel, string(eventType), err == nil, latency)

	return err
}

// PublishBroadcast publishes to all workers
func (p *RedisPublisher) PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error {
	event := events.NewEvent(eventType, message, "")
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	start := time.Now()
	err = p.client.Publish(ctx, events.ChannelBroadcast, data)
	latency := time.Since(start).Milliseconds()
	p.metrics.RecordRedisPublish(events.ChannelBroadcast, string(eventType), err == nil, latency)

	return err
}

// PublishToAgent publishes directly to an agent
func (p *RedisPublisher) PublishToAgent(ctx context.Context, agentID string, data []byte) error {
	channel := events.ChannelForAgent(agentID)

	start := time.Now()
	err := p.client.Publish(ctx, channel, data)
	latency := time.Since(start).Milliseconds()
	p.metrics.RecordRedisPublish(channel, "direct_message", err == nil, latency)

	return err
}

// NoOpPublisher is a no-op implementation for testing or when Redis is disabled
type NoOpPublisher struct{}

func (p *NoOpPublisher) PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error {
	return nil
}

func (p *NoOpPublisher) PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error {
	return nil
}

func (p *NoOpPublisher) PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error {
	return nil
}

func (p *NoOpPublisher) PublishToAgent(ctx context.Context, agentID string, data []byte) error {
	return nil
}

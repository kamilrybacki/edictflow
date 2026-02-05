package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	// Maximum requests allowed in the window
	MaxRequests int
	// Time window for rate limiting
	Window time.Duration
	// Key prefix for Redis storage
	KeyPrefix string
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests: 100,
		Window:      time.Minute,
		KeyPrefix:   "ratelimit",
	}
}

// AuthRateLimitConfig returns stricter limits for auth endpoints
func AuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyPrefix:   "ratelimit:auth",
	}
}

// RateLimiter provides Redis-based rate limiting
type RateLimiter struct {
	redis  *redis.Client
	config RateLimitConfig
}

// NewRateLimiter creates a rate limiter with the given Redis client
func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: config,
	}
}

// Middleware returns HTTP middleware that enforces rate limits
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client identifier (IP or user ID if authenticated)
		identifier := rl.getIdentifier(r)
		key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)

		ctx := r.Context()
		allowed, remaining, resetAt, err := rl.checkLimit(ctx, key)

		if err != nil {
			// On Redis error, allow request but log
			// Fail open to avoid blocking legitimate traffic
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.MaxRequests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(resetAt-time.Now().Unix(), 10))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIdentifier returns a unique identifier for rate limiting
func (rl *RateLimiter) getIdentifier(r *http.Request) string {
	// Check if user is authenticated (use user ID)
	if userID := GetUserID(r.Context()); userID != "" {
		return "user:" + userID
	}

	// Fall back to IP address
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	return "ip:" + ip
}

// checkLimit checks if the request is within rate limits using sliding window
func (rl *RateLimiter) checkLimit(ctx context.Context, key string) (allowed bool, remaining int, resetAt int64, err error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)
	resetAt = now.Add(rl.config.Window).Unix()

	// Use Redis pipeline for atomic operations
	pipe := rl.redis.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiry on the key
	pipe.Expire(ctx, key, rl.config.Window+time.Second)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, resetAt, err
	}

	count := int(countCmd.Val())
	remaining = rl.config.MaxRequests - count - 1
	if remaining < 0 {
		remaining = 0
	}

	allowed = count < rl.config.MaxRequests

	return allowed, remaining, resetAt, nil
}

// RateLimitByPath creates middleware with different limits per path pattern
type RateLimitByPath struct {
	redis    *redis.Client
	limiters map[string]*RateLimiter
	fallback *RateLimiter
}

// NewRateLimitByPath creates a path-based rate limiter
func NewRateLimitByPath(redisClient *redis.Client) *RateLimitByPath {
	return &RateLimitByPath{
		redis:    redisClient,
		limiters: make(map[string]*RateLimiter),
		fallback: NewRateLimiter(redisClient, DefaultRateLimitConfig()),
	}
}

// AddPath adds rate limiting for a specific path prefix
func (rlp *RateLimitByPath) AddPath(pathPrefix string, config RateLimitConfig) *RateLimitByPath {
	rlp.limiters[pathPrefix] = NewRateLimiter(rlp.redis, config)
	return rlp
}

// Middleware returns HTTP middleware that applies path-specific rate limits
func (rlp *RateLimitByPath) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Find matching limiter
		var limiter *RateLimiter
		for prefix, l := range rlp.limiters {
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				limiter = l
				break
			}
		}

		if limiter == nil {
			limiter = rlp.fallback
		}

		// Apply the limiter's middleware
		limiter.Middleware(next).ServeHTTP(w, r)
	})
}

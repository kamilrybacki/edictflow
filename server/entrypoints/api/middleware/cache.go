package middleware

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/common/workerpool"
)

// CacheConfig defines caching parameters
type CacheConfig struct {
	// Time-to-live for cached responses
	TTL time.Duration
	// Key prefix for Redis storage
	KeyPrefix string
	// Only cache these status codes (empty = cache all 2xx)
	CacheableCodes []int
}

// DefaultCacheConfig returns sensible defaults
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:            60 * time.Second,
		KeyPrefix:      "cache:api",
		CacheableCodes: []int{200},
	}
}

// Cache provides Redis-based response caching
type Cache struct {
	redis  *redisAdapter.Client
	config CacheConfig
}

// NewCache creates a cache middleware with the given Redis client
func NewCache(redisClient *redisAdapter.Client, config CacheConfig) *Cache {
	return &Cache{
		redis:  redisClient,
		config: config,
	}
}

// cachedResponse represents a cached HTTP response
type cachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	CachedAt   int64             `json:"cached_at"`
}

// Middleware returns HTTP middleware that caches GET responses
func (c *Cache) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		cacheKey := c.buildCacheKey(r)

		// Try to get from cache
		if cached, ok := c.getFromCache(ctx, cacheKey); ok {
			c.writeCachedResponse(w, cached)
			return
		}

		// Capture response
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     200,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(recorder, r)

		// Cache if status code is cacheable (use worker pool to avoid unbounded goroutines)
		if c.isCacheable(recorder.statusCode) {
			// Capture values for closure
			key := cacheKey
			body := recorder.body.Bytes()
			statusCode := recorder.statusCode
			headers := make(map[string]string)
			for k, v := range recorder.Header() {
				if len(v) > 0 && shouldCacheHeader(k) {
					headers[k] = v[0]
				}
			}
			workerpool.DefaultCachePool.Submit(func() {
				c.saveToCacheData(context.Background(), key, statusCode, headers, body)
			})
		}
	})
}

// buildCacheKey creates a unique cache key from the request
func (c *Cache) buildCacheKey(r *http.Request) string {
	// Include path, query params, and user ID for personalized caching
	parts := []string{
		r.URL.Path,
		r.URL.RawQuery,
	}

	// Include user ID if authenticated (for user-specific caching)
	if userID := GetUserID(r.Context()); userID != "" {
		parts = append(parts, "user:"+userID)
	}

	// Include team ID if present
	if teamID := GetTeamID(r.Context()); teamID != "" {
		parts = append(parts, "team:"+teamID)
	}

	// Use FNV-1a hash (faster than SHA256 for non-crypto use)
	h := fnv.New64a()
	h.Write([]byte(strings.Join(parts, "|")))
	return c.config.KeyPrefix + ":" + hex.EncodeToString(h.Sum(nil))
}

// getFromCache retrieves a cached response
func (c *Cache) getFromCache(ctx context.Context, key string) (*cachedResponse, bool) {
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var cached cachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	return &cached, true
}

// saveToCache stores a response in the cache
func (c *Cache) saveToCache(ctx context.Context, key string, recorder *responseRecorder) {
	headers := make(map[string]string)
	for k, v := range recorder.Header() {
		if len(v) > 0 && shouldCacheHeader(k) {
			headers[k] = v[0]
		}
	}
	c.saveToCacheData(ctx, key, recorder.statusCode, headers, recorder.body.Bytes())
}

// saveToCacheData stores response data in the cache (used by worker pool)
func (c *Cache) saveToCacheData(ctx context.Context, key string, statusCode int, headers map[string]string, body []byte) {
	cached := cachedResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
		CachedAt:   time.Now().Unix(),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return
	}

	c.redis.Set(ctx, key, data, c.config.TTL)
}

// writeCachedResponse writes a cached response to the client
func (c *Cache) writeCachedResponse(w http.ResponseWriter, cached *cachedResponse) {
	// Set cached headers
	for k, v := range cached.Headers {
		w.Header().Set(k, v)
	}

	// Add cache indicator header
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("X-Cache-Age", time.Since(time.Unix(cached.CachedAt, 0)).String())

	w.WriteHeader(cached.StatusCode)
	w.Write(cached.Body)
}

// isCacheable checks if a status code should be cached
func (c *Cache) isCacheable(statusCode int) bool {
	if len(c.config.CacheableCodes) == 0 {
		return statusCode >= 200 && statusCode < 300
	}

	for _, code := range c.config.CacheableCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// shouldCacheHeader returns true if the header should be preserved in cache
func shouldCacheHeader(name string) bool {
	name = strings.ToLower(name)
	switch name {
	case "content-type", "content-language", "content-encoding":
		return true
	default:
		return false
	}
}

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// CacheInvalidator provides cache invalidation capabilities
type CacheInvalidator struct {
	redis     *redisAdapter.Client
	keyPrefix string
}

// NewCacheInvalidator creates a cache invalidator
func NewCacheInvalidator(redisClient *redisAdapter.Client, keyPrefix string) *CacheInvalidator {
	return &CacheInvalidator{
		redis:     redisClient,
		keyPrefix: keyPrefix,
	}
}

// InvalidatePattern invalidates all cache keys matching a pattern
func (ci *CacheInvalidator) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := ci.keyPrefix + ":" + pattern + "*"

	// Use SCAN to find matching keys (safer than KEYS for production)
	var cursor uint64
	var keysToDelete []string

	for {
		keys, nextCursor, err := ci.redis.Underlying().Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return err
		}

		keysToDelete = append(keysToDelete, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keysToDelete) > 0 {
		return ci.redis.Del(ctx, keysToDelete...)
	}

	return nil
}

// InvalidateRules invalidates all rule-related caches
func (ci *CacheInvalidator) InvalidateRules(ctx context.Context) error {
	return ci.InvalidatePattern(ctx, "rules")
}

// InvalidateTeams invalidates all team-related caches
func (ci *CacheInvalidator) InvalidateTeams(ctx context.Context) error {
	return ci.InvalidatePattern(ctx, "teams")
}

// InvalidateForTeam invalidates caches for a specific team
func (ci *CacheInvalidator) InvalidateForTeam(ctx context.Context, teamID string) error {
	return ci.InvalidatePattern(ctx, "team:"+teamID)
}

package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type PermissionProvider interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

// permissionCacheEntry holds cached permissions with expiration
type permissionCacheEntry struct {
	permissions []string
	expiresAt   time.Time
}

type Permission struct {
	provider PermissionProvider
	cache    map[string]permissionCacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
}

func NewPermission(provider PermissionProvider) *Permission {
	return &Permission{
		provider: provider,
		cache:    make(map[string]permissionCacheEntry),
		cacheTTL: 5 * time.Minute, // Default 5 minute TTL
	}
}

// NewPermissionWithTTL creates a Permission middleware with custom cache TTL
func NewPermissionWithTTL(provider PermissionProvider, ttl time.Duration) *Permission {
	return &Permission{
		provider: provider,
		cache:    make(map[string]permissionCacheEntry),
		cacheTTL: ttl,
	}
}

// InvalidateCache removes a user's cached permissions (call on role changes)
func (p *Permission) InvalidateCache(userID string) {
	p.cacheMu.Lock()
	delete(p.cache, userID)
	p.cacheMu.Unlock()
}

// InvalidateAllCache clears the entire permission cache
func (p *Permission) InvalidateAllCache() {
	p.cacheMu.Lock()
	p.cache = make(map[string]permissionCacheEntry)
	p.cacheMu.Unlock()
}

// RequirePermission returns middleware that requires the user to have a specific permission
func (p *Permission) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			permissions, err := p.getUserPermissions(r.Context(), userID)
			if err != nil {
				http.Error(w, "failed to get permissions", http.StatusInternalServerError)
				return
			}

			if !hasPermission(permissions, permission) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that requires the user to have at least one of the specified permissions
func (p *Permission) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userPermissions, err := p.getUserPermissions(r.Context(), userID)
			if err != nil {
				http.Error(w, "failed to get permissions", http.StatusInternalServerError)
				return
			}

			for _, required := range permissions {
				if hasPermission(userPermissions, required) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

// RequireAllPermissions returns middleware that requires the user to have all of the specified permissions
func (p *Permission) RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userPermissions, err := p.getUserPermissions(r.Context(), userID)
			if err != nil {
				http.Error(w, "failed to get permissions", http.StatusInternalServerError)
				return
			}

			for _, required := range permissions {
				if !hasPermission(userPermissions, required) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getUserPermissions checks context first (from JWT), then cache, then provider
func (p *Permission) getUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// Fast path: permissions from JWT in context
	if perms := GetPermissions(ctx); perms != nil {
		return perms, nil
	}

	// Check in-memory cache
	p.cacheMu.RLock()
	if entry, ok := p.cache[userID]; ok && time.Now().Before(entry.expiresAt) {
		p.cacheMu.RUnlock()
		return entry.permissions, nil
	}
	p.cacheMu.RUnlock()

	// Slow path: fetch from provider (database)
	perms, err := p.provider.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	p.cacheMu.Lock()
	p.cache[userID] = permissionCacheEntry{
		permissions: perms,
		expiresAt:   time.Now().Add(p.cacheTTL),
	}
	p.cacheMu.Unlock()

	return perms, nil
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}

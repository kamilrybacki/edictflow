package middleware

import (
	"context"
	"net/http"
)

type PermissionProvider interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

type Permission struct {
	provider PermissionProvider
}

func NewPermission(provider PermissionProvider) *Permission {
	return &Permission{provider: provider}
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

			permissions, err := p.provider.GetUserPermissions(r.Context(), userID)
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

			userPermissions, err := p.provider.GetUserPermissions(r.Context(), userID)
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

			userPermissions, err := p.provider.GetUserPermissions(r.Context(), userID)
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

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}

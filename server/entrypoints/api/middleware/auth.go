package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey      contextKey = "user_id"
	teamIDKey      contextKey = "team_id"
	permissionsKey contextKey = "permissions"
)

// Exported keys for testing
var (
	UserIDContextKey      = userIDKey
	TeamIDContextKey      = teamIDKey
	PermissionsContextKey = permissionsKey
)

type Auth struct {
	secret []byte
}

func NewAuth(secret string) *Auth {
	return &Auth{secret: []byte(secret)}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return a.secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		if sub, ok := claims["sub"].(string); ok {
			ctx = context.WithValue(ctx, userIDKey, sub)
		}
		if teamID, ok := claims["team_id"].(string); ok {
			ctx = context.WithValue(ctx, teamIDKey, teamID)
		}
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			perms := make([]string, 0, len(permissions))
			for _, p := range permissions {
				if s, ok := p.(string); ok {
					perms = append(perms, s)
				}
			}
			ctx = context.WithValue(ctx, permissionsKey, perms)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware extracts user info from token if present but doesn't block unauthenticated requests
func (a *Auth) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header - proceed without user context
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format - proceed without user context
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return a.secret, nil
		})

		if err != nil || !token.Valid {
			// Invalid token - proceed without user context
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		if sub, ok := claims["sub"].(string); ok {
			ctx = context.WithValue(ctx, userIDKey, sub)
		}
		if teamID, ok := claims["team_id"].(string); ok {
			ctx = context.WithValue(ctx, teamIDKey, teamID)
		}
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			perms := make([]string, 0, len(permissions))
			for _, p := range permissions {
				if s, ok := p.(string); ok {
					perms = append(perms, s)
				}
			}
			ctx = context.WithValue(ctx, permissionsKey, perms)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func GetTeamID(ctx context.Context) string {
	if v := ctx.Value(teamIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func GetPermissions(ctx context.Context) []string {
	if v := ctx.Value(permissionsKey); v != nil {
		return v.([]string)
	}
	return nil
}

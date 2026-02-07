package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/metrics"
	"github.com/kamilrybacki/edictflow/server/services/publisher"
)

// FullAuditService combines the read and write interfaces for audit logging
type FullAuditService interface {
	handlers.AuditService
	handlers.RuleAuditLogger
}

type Config struct {
	JWTSecret                  string
	BaseURL                    string
	TeamService                handlers.TeamService
	RuleService                handlers.RuleService
	CategoryService            handlers.CategoryService
	ChangeService              handlers.ChangeService
	ExceptionService           handlers.ExceptionService
	NotificationService        handlers.NotificationService
	NotificationChannelService handlers.NotificationChannelService
	DeviceAuthService          handlers.DeviceAuthService
	AuthService                handlers.AuthService
	UserService                handlers.UserService
	UsersService               handlers.UsersService
	ApprovalsService           handlers.ApprovalsService
	InviteService              handlers.InviteService
	AuditService               FullAuditService
	PermissionProvider         middleware.PermissionProvider
	Publisher                  publisher.Publisher
	MetricsService             metrics.Service
	RedisClient                *redisAdapter.Client
}

func NewRouter(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	// Metrics middleware (first in chain to capture all requests)
	if cfg.MetricsService != nil {
		metricsMiddleware := middleware.NewMetrics(cfg.MetricsService)
		r.Use(metricsMiddleware.Middleware)
	}

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	auth := middleware.NewAuth(cfg.JWTSecret)
	perm := middleware.NewPermission(cfg.PermissionProvider)

	// Rate limiting and caching (only if Redis is available)
	var rateLimiter *middleware.RateLimitByPath
	var cache *middleware.Cache
	var cacheInvalidator *middleware.CacheInvalidator

	if cfg.RedisClient != nil {
		// Rate limiting with different limits per path
		rateLimiter = middleware.NewRateLimitByPath(cfg.RedisClient.Underlying())
		rateLimiter.AddPath("/api/v1/auth", middleware.AuthRateLimitConfig())
		rateLimiter.AddPath("/api/v1", middleware.DefaultRateLimitConfig())

		// Caching for GET requests
		cache = middleware.NewCache(cfg.RedisClient, middleware.CacheConfig{
			TTL:            60 * time.Second,
			KeyPrefix:      "cache:api",
			CacheableCodes: []int{200},
		})

		// Cache invalidator for write operations
		cacheInvalidator = middleware.NewCacheInvalidator(cfg.RedisClient, "cache:api")
		_ = cacheInvalidator // Used by handlers
	}

	// Root endpoint (public)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"edictflow","version":"1.0.0","status":"running"}`))
	})

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Auth routes (mix of public and protected)
	r.Route("/api/v1/auth", func(r chi.Router) {
		// Apply rate limiting to auth routes (stricter limits)
		if rateLimiter != nil {
			r.Use(rateLimiter.Middleware)
		}

		// Public endpoints - login and register don't require auth
		if cfg.AuthService != nil && cfg.UserService != nil {
			authHandler := handlers.NewAuthHandler(cfg.AuthService, cfg.UserService)
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)

			// Protected auth endpoints
			r.Group(func(r chi.Router) {
				r.Use(auth.Middleware)
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.GetProfile)
				r.Put("/me", authHandler.UpdateProfile)
				r.Put("/me/password", authHandler.UpdatePassword)
			})
		}

		// Device auth routes (public - no auth required for device code flow)
		if cfg.DeviceAuthService != nil {
			deviceAuthHandler := handlers.NewDeviceAuthHandler(cfg.DeviceAuthService, cfg.BaseURL)
			r.Post("/device", deviceAuthHandler.InitiateDeviceAuth)
			r.Post("/device/token", deviceAuthHandler.PollForToken)
		}
	})

	// User-facing device verification page (separate from API)
	// Auth check is done inside handler to enable redirect to login
	if cfg.DeviceAuthService != nil {
		deviceAuthHandler := handlers.NewDeviceAuthHandler(cfg.DeviceAuthService, cfg.BaseURL)
		r.Route("/auth/device", func(r chi.Router) {
			// Use optional auth - extracts user if token present but doesn't block
			r.Use(auth.OptionalMiddleware)
			r.Get("/verify", deviceAuthHandler.VerifyPage)
			r.Post("/verify", deviceAuthHandler.VerifyPage)
		})
	}

	// API routes (protected)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware)

		// Apply rate limiting to all API routes
		if rateLimiter != nil {
			r.Use(rateLimiter.Middleware)
		}

		r.Route("/teams", func(r chi.Router) {
			// Cache GET requests for teams
			if cache != nil {
				r.Use(cache.Middleware)
			}
			h := handlers.NewTeamsHandler(cfg.TeamService)
			h.RegisterRoutes(r)
		})

		r.Route("/rules", func(r chi.Router) {
			// Cache GET requests for rules
			if cache != nil {
				r.Use(cache.Middleware)
			}
			var h *handlers.RulesHandler
			if cfg.AuditService != nil {
				h = handlers.NewRulesHandlerWithAudit(cfg.RuleService, cfg.Publisher, cfg.AuditService)
			} else {
				h = handlers.NewRulesHandler(cfg.RuleService, cfg.Publisher)
			}
			// Wire up user lookup for resolving createdBy names
			if cfg.UsersService != nil {
				h = h.WithUserLookup(cfg.UsersService)
			}
			h.RegisterRoutes(r)
		})

		if cfg.CategoryService != nil {
			r.Route("/categories", func(r chi.Router) {
				h := handlers.NewCategoriesHandler(cfg.CategoryService)
				h.RegisterRoutes(r)
			})
		}

		r.Route("/changes", func(r chi.Router) {
			r.Use(perm.RequirePermission("changes.view"))
			h := handlers.NewChangesHandler(cfg.ChangeService)
			r.Get("/", h.List)
			r.Get("/{id}", h.Get)
			r.Group(func(r chi.Router) {
				r.Use(perm.RequirePermission("changes.approve"))
				r.Post("/{id}/approve", h.Approve)
				r.Post("/{id}/reject", h.Reject)
			})
		})

		r.Route("/exceptions", func(r chi.Router) {
			r.Use(perm.RequirePermission("exceptions.view"))
			h := handlers.NewExceptionsHandler(cfg.ExceptionService)
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Group(func(r chi.Router) {
				r.Use(perm.RequirePermission("exceptions.approve"))
				r.Post("/{id}/approve", h.Approve)
				r.Post("/{id}/deny", h.Deny)
			})
		})

		if cfg.ApprovalsService != nil {
			r.Route("/approvals", func(r chi.Router) {
				h := handlers.NewApprovalsHandler(cfg.ApprovalsService)
				h.RegisterRoutes(r)
			})
		}

		r.Route("/notifications", func(r chi.Router) {
			h := handlers.NewNotificationsHandler(cfg.NotificationService)
			h.RegisterRoutes(r)
		})

		r.Route("/notification-channels", func(r chi.Router) {
			r.Use(perm.RequirePermission("notifications.manage"))
			h := handlers.NewNotificationChannelsHandler(cfg.NotificationChannelService)
			h.RegisterRoutes(r)
		})

		// Users routes
		if cfg.UsersService != nil {
			r.Route("/users", func(r chi.Router) {
				h := handlers.NewUsersHandler(cfg.UsersService)
				h.RegisterRoutes(r)
			})
		}

		// Invite join route (authenticated, but not team-specific)
		if cfg.InviteService != nil {
			r.Route("/invites", func(r chi.Router) {
				h := handlers.NewInvitesHandler(cfg.InviteService)
				h.RegisterRoutes(r)
			})
		}

		// Audit log routes
		if cfg.AuditService != nil {
			r.Route("/audit", func(r chi.Router) {
				h := handlers.NewAuditHandler(cfg.AuditService)
				h.RegisterRoutes(r)
			})
		}
	})

	return r
}

package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
)

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
	ApprovalsService           handlers.ApprovalsService
	PermissionProvider         middleware.PermissionProvider
	Publisher                  publisher.Publisher
}

func NewRouter(cfg Config) *chi.Mux {
	r := chi.NewRouter()

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

	// Root endpoint (public)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"claudeception","version":"1.0.0","status":"running"}`))
	})

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Auth routes (mix of public and protected)
	r.Route("/api/v1/auth", func(r chi.Router) {
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

			// Verification page requires authentication
			r.Group(func(r chi.Router) {
				r.Use(auth.Middleware)
				r.Get("/device/verify", deviceAuthHandler.VerifyPage)
				r.Post("/device/verify", deviceAuthHandler.VerifyPage)
			})
		}
	})

	// API routes (protected)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware)

		r.Route("/teams", func(r chi.Router) {
			h := handlers.NewTeamsHandler(cfg.TeamService)
			h.RegisterRoutes(r)
		})

		r.Route("/rules", func(r chi.Router) {
			h := handlers.NewRulesHandler(cfg.RuleService, cfg.Publisher)
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
	})

	return r
}

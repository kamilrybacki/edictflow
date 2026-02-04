package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
)

type Config struct {
	JWTSecret                  string
	TeamService                handlers.TeamService
	RuleService                handlers.RuleService
	ChangeService              handlers.ChangeService
	ExceptionService           handlers.ExceptionService
	NotificationService        handlers.NotificationService
	NotificationChannelService handlers.NotificationChannelService
	PermissionProvider         middleware.PermissionProvider
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

	// API routes (protected)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware)

		r.Route("/teams", func(r chi.Router) {
			h := handlers.NewTeamsHandler(cfg.TeamService)
			h.RegisterRoutes(r)
		})

		r.Route("/rules", func(r chi.Router) {
			h := handlers.NewRulesHandler(cfg.RuleService)
			h.RegisterRoutes(r)
		})

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

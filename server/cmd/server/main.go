package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kamilrybacki/claudeception/server/adapters/postgres"
	"github.com/kamilrybacki/claudeception/server/configurator"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api"
	"github.com/kamilrybacki/claudeception/server/entrypoints/ws"
)

func main() {
	settings := configurator.LoadSettings()
	ctx := context.Background()

	// Initialize database connection
	pool, err := postgres.NewPool(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize repositories
	teamDB := postgres.NewTeamDB(pool)
	ruleDB := postgres.NewRuleDB(pool)

	// Create services that implement the handler interfaces
	teamService := &teamServiceImpl{db: teamDB}
	ruleService := &ruleServiceImpl{db: ruleDB}

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create router with real services
	router := api.NewRouter(api.Config{
		JWTSecret:   settings.JWTSecret,
		TeamService: teamService,
		RuleService: ruleService,
	})

	// Add WebSocket endpoint
	wsHandler := ws.NewHandler(hub, nil)
	router.Get("/ws", wsHandler.ServeHTTP)

	server := &http.Server{
		Addr:         ":" + settings.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		close(done)
	}()

	log.Printf("Server starting on port %s", settings.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	<-done
	log.Println("Server stopped")
}

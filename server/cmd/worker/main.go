package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/configurator"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/claudeception/server/worker"
)

func main() {
	settings := configurator.LoadSettings()
	ctx := context.Background()

	// Initialize Redis
	redisClient, err := redisAdapter.NewClient(settings.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize worker hub
	hub := worker.NewHub(redisClient)
	go hub.Run()

	// Create router
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	// Health endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		agents, teams, subs := hub.Stats()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","agents":` + strconv.Itoa(agents) + `,"teams":` + strconv.Itoa(teams) + `,"subscriptions":` + strconv.Itoa(subs) + `}`))
	})

	// WebSocket endpoint with auth
	auth := middleware.NewAuth(settings.JWTSecret)
	wsHandler := worker.NewHandler(hub)
	router.With(auth.Middleware).Get("/ws", wsHandler.ServeHTTP)

	// Get worker port (different from API port)
	port := getEnv("WORKER_PORT", "8081")

	server := &http.Server{
		Addr:         ":" + port,
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
		log.Println("Shutting down worker...")

		hub.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Worker forced to shutdown: %v", err)
		}

		close(done)
	}()

	log.Printf("Worker starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Worker error: %v", err)
	}

	<-done
	log.Println("Worker stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/adapters/splunk"
	"github.com/kamilrybacki/edictflow/server/configurator"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/metrics"
	"github.com/kamilrybacki/edictflow/server/worker"
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

	// Initialize metrics service
	var metricsService metrics.Service
	if settings.SplunkEnabled && settings.SplunkHECURL != "" {
		hostname, _ := os.Hostname()
		metricsService = metrics.NewSplunkService(metrics.Config{
			SplunkConfig: splunk.Config{
				HECURL:        settings.SplunkHECURL,
				Token:         settings.SplunkHECToken,
				Source:        settings.SplunkSource,
				SourceType:    settings.SplunkSourceType,
				Index:         settings.SplunkIndex,
				SkipTLSVerify: settings.SplunkSkipTLSVerify,
			},
			Hostname: hostname + "-worker",
		})
		defer metricsService.Close()
		log.Println("Splunk metrics enabled")
	} else {
		metricsService = &metrics.NoOpService{}
	}

	// Initialize worker hub
	hub := worker.NewHub(redisClient)
	hub.SetMetrics(metricsService)
	go hub.Run()

	// Start hub stats reporter and worker heartbeat if metrics enabled
	if settings.SplunkEnabled {
		workerID, _ := os.Hostname()
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					agents, teams, subs := hub.Stats()
					metricsService.RecordHubStats(agents, teams, subs)
					metricsService.RecordWorkerHeartbeat(workerID, agents, teams)

					// Record Redis health check
					start := time.Now()
					err := redisClient.Ping(context.Background())
					latency := time.Since(start).Milliseconds()
					status := "healthy"
					if err != nil {
						status = "unhealthy"
					}
					metricsService.RecordHealthCheck("redis", status, latency)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

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

	// Agents list endpoint (for admin UI)
	router.Get("/agents", func(w http.ResponseWriter, r *http.Request) {
		allAgents := hub.ListAgents()

		// Filter by team_id if provided
		teamID := r.URL.Query().Get("team_id")
		var agents []worker.AgentInfo
		if teamID != "" {
			for _, a := range allAgents {
				if a.TeamID == teamID {
					agents = append(agents, a)
				}
			}
		} else {
			agents = allAgents
		}

		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(agents)
		w.Write(data)
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

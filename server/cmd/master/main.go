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
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/configurator"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api"
	"github.com/kamilrybacki/claudeception/server/services/approvals"
	"github.com/kamilrybacki/claudeception/server/services/auth"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
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

	// Initialize Redis (optional - graceful degradation)
	var pub publisher.Publisher
	redisClient, err := redisAdapter.NewClient(settings.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis not available, events will not be published: %v", err)
		pub = &publisher.NoOpPublisher{}
	} else {
		defer redisClient.Close()
		if err := redisClient.Ping(ctx); err != nil {
			log.Printf("Warning: Redis ping failed: %v", err)
			pub = &publisher.NoOpPublisher{}
		} else {
			log.Println("Connected to Redis")
			pub = publisher.NewRedisPublisher(redisClient)
		}
	}

	// Initialize repositories
	teamDB := postgres.NewTeamDB(pool)
	ruleDB := postgres.NewRuleDB(pool)
	categoryDB := postgres.NewCategoryDB(pool)
	userDB := postgres.NewUserDB(pool)
	roleDB := postgres.NewRoleDB(pool)
	approvalDB := postgres.NewRuleApprovalDB(pool)
	approvalConfigDB := postgres.NewApprovalConfigDB(pool)

	// Create services that implement the handler interfaces
	teamService := &teamServiceImpl{db: teamDB}
	ruleService := &ruleServiceImpl{db: ruleDB, categoryDB: categoryDB}
	categoryService := &categoryServiceImpl{db: categoryDB}
	userService := &userServiceImpl{db: userDB}
	authService := auth.NewService(userDB, roleDB, settings.JWTSecret, 24*time.Hour)
	approvalsService := approvals.NewService(ruleDB, approvalDB, approvalConfigDB, roleDB)

	// Create router (no WebSocket - handled by workers)
	router := api.NewRouter(api.Config{
		JWTSecret:        settings.JWTSecret,
		TeamService:      teamService,
		RuleService:      ruleService,
		CategoryService:  categoryService,
		AuthService:      authService,
		UserService:      userService,
		ApprovalsService: approvalsService,
		Publisher:        pub,
	})

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
		log.Println("Shutting down master...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Master forced to shutdown: %v", err)
		}

		close(done)
	}()

	log.Printf("Master API starting on port %s", settings.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Master error: %v", err)
	}

	<-done
	log.Println("Master stopped")
}

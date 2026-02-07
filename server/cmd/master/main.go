package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kamilrybacki/edictflow/server/adapters/postgres"
	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/adapters/splunk"
	"github.com/kamilrybacki/edictflow/server/configurator"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api"
	"github.com/kamilrybacki/edictflow/server/services/approvals"
	"github.com/kamilrybacki/edictflow/server/services/audit"
	"github.com/kamilrybacki/edictflow/server/services/auth"
	"github.com/kamilrybacki/edictflow/server/services/deviceauth"
	"github.com/kamilrybacki/edictflow/server/services/metrics"
	"github.com/kamilrybacki/edictflow/server/services/notifications"
	"github.com/kamilrybacki/edictflow/server/services/publisher"
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
			Hostname: hostname + "-master",
		})
		defer metricsService.Close()
		log.Println("Splunk metrics enabled")
	} else {
		metricsService = &metrics.NoOpService{}
	}

	// Initialize Redis (optional - graceful degradation)
	var pub publisher.Publisher
	var redisClient *redisAdapter.Client
	redisClient, err = redisAdapter.NewClient(settings.RedisURL)
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
			redisPub := publisher.NewRedisPublisher(redisClient)
			redisPub.SetMetrics(metricsService)
			pub = redisPub
		}
	}

	// Start health metrics reporter if enabled
	if settings.SplunkEnabled {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// Record database pool stats
					stat := pool.Stat()
					metricsService.RecordDBPoolStats(stat.TotalConns(), stat.AcquiredConns(), stat.IdleConns(), stat.MaxConns())

					// Record database health check
					dbStart := time.Now()
					dbErr := pool.Ping(context.Background())
					dbLatency := time.Since(dbStart).Milliseconds()
					dbStatus := "healthy"
					if dbErr != nil {
						dbStatus = "unhealthy"
					}
					metricsService.RecordHealthCheck("postgres", dbStatus, dbLatency)

					// Record Redis health check if available
					if redisClient != nil {
						redisStart := time.Now()
						redisErr := redisClient.Ping(context.Background())
						redisLatency := time.Since(redisStart).Milliseconds()
						redisStatus := "healthy"
						if redisErr != nil {
							redisStatus = "unhealthy"
						}
						metricsService.RecordHealthCheck("redis", redisStatus, redisLatency)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Initialize repositories
	teamDB := postgres.NewTeamDB(pool)
	teamInviteDB := postgres.NewTeamInviteDB(pool)
	ruleDB := postgres.NewRuleDB(pool)
	categoryDB := postgres.NewCategoryDB(pool)
	userDB := postgres.NewUserDB(pool)
	roleDB := postgres.NewRoleDB(pool)
	approvalDB := postgres.NewRuleApprovalDB(pool)
	approvalConfigDB := postgres.NewApprovalConfigDB(pool)
	deviceCodeDB := postgres.NewDeviceCodeDB(pool)
	notificationDB := postgres.NewNotificationDB(pool)
	notificationChannelDB := postgres.NewNotificationChannelDB(pool)
	auditDB := postgres.NewAuditDB(pool)

	// Create services that implement the handler interfaces
	teamService := &teamServiceImpl{db: teamDB, inviteDB: teamInviteDB, userDB: userDB}
	ruleService := &ruleServiceImpl{db: ruleDB, categoryDB: categoryDB}
	categoryService := &categoryServiceImpl{db: categoryDB}
	userService := &userServiceImpl{db: userDB}
	usersService := &usersServiceImpl{db: userDB}
	authService := auth.NewService(userDB, roleDB, settings.JWTSecret, 24*time.Hour)
	auditService := audit.NewService(auditDB)
	approvalsService := approvals.NewService(ruleDB, approvalDB, approvalConfigDB, roleDB).WithAuditLogger(auditService)
	deviceAuthService := deviceauth.NewService(deviceCodeDB, authService)
	notificationSvc := notifications.NewService(notificationDB, notificationChannelDB)
	notificationService := &notificationServiceWrapper{svc: notificationSvc}

	// Create router (no WebSocket - handled by workers)
	router := api.NewRouter(api.Config{
		JWTSecret:           settings.JWTSecret,
		BaseURL:             settings.BaseURL,
		TeamService:         teamService,
		RuleService:         ruleService,
		CategoryService:     categoryService,
		AuthService:         authService,
		UserService:         userService,
		UsersService:        usersService,
		ApprovalsService:    approvalsService,
		DeviceAuthService:   deviceAuthService,
		NotificationService: notificationService,
		InviteService:       teamService,
		AuditService:        auditService,
		Publisher:           pub,
		MetricsService:      metricsService,
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

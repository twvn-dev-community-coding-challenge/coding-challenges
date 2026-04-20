package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dotdak/sms-otp/handlers"
	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/httpmw"
	"github.com/dotdak/sms-otp/internal/metrics"
	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/dotdak/sms-otp/worker"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}
	obslog.Init()

	// Initialize worker context
	ctx := context.Background()

	// Initialize Postgres database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalln("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-Migrate the database schema
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate models: %v", err)
	}
	if err := repository.AutoMigrateSMS(db); err != nil {
		log.Fatalf("Failed to auto-migrate SMS tables: %v", err)
	}

	// Instantiate the User repository
	userRepo := repository.NewUserRepository(db)

	// Get Redis URL and Queue Name from environment variables, fallback to defaults
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "sms_queue"
	}

	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(httpmw.CorrelationID())
	e.Use(httpmw.RequestID())
	e.Use(httpmw.AccessLog())
	e.Use(metrics.EchoMiddleware())

	// CORS configuration for React frontend
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "https://hoppscotch.io"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
			httpmw.HeaderXCorrelationID,
		},
	}))

	// Initialize SMS Service
	smsRepo := repository.NewSMSGormRepository(db)
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	smsAdapters := map[providers.Provider]providers.SMSProvider{
		providers.ProviderTwilio:      providers.NewTwilioAdapter(os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN")),
		providers.ProviderVonage:      providers.NewMockProvider(providers.ProviderVonage),
		providers.ProviderInfobip:     providers.NewMockProvider(providers.ProviderInfobip),
		providers.ProviderAWSSNS:      providers.NewMockProvider(providers.ProviderAWSSNS),
		providers.ProviderTelnyx:      providers.NewMockProvider(providers.ProviderTelnyx),
		providers.ProviderMessageBird: providers.NewMockProvider(providers.ProviderMessageBird),
		providers.ProviderSinch:       providers.NewMockProvider(providers.ProviderSinch),
	}
	router.RegisterDefaultRoutes(smsAdapters)

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL (required for SMS queue and rate limiting): %v", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed (required for SMS queue and rate limiting): %v", err)
	}
	limiter := ratelimit.New(rdb)

	smsQueue, err := worker.NewSMSQueue(ctx, rdb, queueName)
	if err != nil {
		log.Fatalf("failed to create BullMQ SMS queue: %v", err)
	}
	smsPublisher := sms.NewBullMQSendPublisher(smsQueue)

	smsService := sms.NewSMSService(smsRepo, resolver, router, limiter, smsPublisher)

	go func() {
		if err := worker.StartSMSWorker(ctx, rdb, queueName, smsService); err != nil {
			log.Printf("Failed to start BullMQ SMS worker: %v", err)
		}
	}()

	costTracker := sms.NewInMemoryCostTracker()
	smsService.RegisterGlobalObserver(costTracker)
	metrics.WireSMSTelemetry()
	metrics.WireCostTelemetry(costTracker)
	statusObs := metrics.NewSMSStatusObserver()
	smsService.RegisterSourceObserver(sms.SendSourceAuth, statusObs)
	smsService.RegisterSourceObserver(sms.SendSourceAPI, statusObs)

	smsHandler := handlers.NewSMSHandler(smsService, costTracker)

	// Initialize Route Handlers
	authHandler := handlers.NewAuthHandler(userRepo, smsService, limiter)

	e.GET("/metrics", echo.WrapHandler(metrics.Handler()))

	// Define routes
	e.POST("/api/register", authHandler.Register)
	e.POST("/api/verify", authHandler.VerifyOTP)
	e.POST("/api/login", authHandler.Login)
	e.POST("/api/login/verify", authHandler.LoginVerify)

	smsInternalKey := os.Getenv("SMS_INTERNAL_API_KEY")
	smsGroup := e.Group("/api/sms")
	if smsInternalKey != "" {
		smsGroup.Use(httpmw.InternalAPIKey(smsInternalKey))
	}
	smsGroup.POST("/send", smsHandler.SendSMS)
	smsGroup.GET("/messages/:id", smsHandler.GetMessage)
	smsGroup.GET("/messages", smsHandler.ListMessages)
	smsGroup.GET("/stats", smsHandler.GetStats)

	callbackSecret := os.Getenv("SMS_CALLBACK_SECRET")
	if callbackSecret != "" {
		e.POST("/api/sms/callback", smsHandler.HandleCallback, httpmw.CallbackSecret(callbackSecret))
	} else {
		e.POST("/api/sms/callback", smsHandler.HandleCallback)
	}

	// Start the server on port 8080
	e.Logger.Fatal(e.Start(":8080"))
}

package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/lppduy/go-asynq-loadtest/internal/config"
	"github.com/lppduy/go-asynq-loadtest/internal/handler"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
	"github.com/lppduy/go-asynq-loadtest/internal/service"
	"github.com/lppduy/go-asynq-loadtest/pkg/database"
)

func main() {
	log.Println("🚀 Starting Order Processing API...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Database configuration
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create Asynq client for enqueueing tasks
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer asynqClient.Close()

	log.Printf("✅ Connected to Redis: %s", cfg.Redis.Addr)

	// Initialize layers (Dependency Injection)
	orderRepo := repository.NewGormOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	taskRetention := time.Duration(cfg.Worker.RetentionMinutes) * time.Minute
	orderHandler := handler.NewOrderHandler(orderService, asynqClient, taskRetention)

	// Setup Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "order-processing-api",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Order endpoints
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)           // Create new order
			orders.GET("", orderHandler.ListOrders)             // List all orders
			orders.GET("/:id", orderHandler.GetOrder)           // Get order by ID
			orders.GET("/:id/status", orderHandler.GetOrderStatus) // Get order status
			orders.POST("/:id/cancel", orderHandler.CancelOrder)   // Cancel order
		}
	}

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("✅ API server running on http://localhost%s", port)
	log.Println("📚 Endpoints:")
	log.Println("   - POST   /api/v1/orders          (Create order)")
	log.Println("   - GET    /api/v1/orders          (List orders)")
	log.Println("   - GET    /api/v1/orders/:id      (Get order)")
	log.Println("   - GET    /api/v1/orders/:id/status (Get status)")
	log.Println("   - POST   /api/v1/orders/:id/cancel (Cancel order)")
	log.Println("")
	log.Printf("💡 Try: curl http://localhost:%s/health", cfg.Server.Port)
	log.Println("")
	log.Println("📋 Background tasks will be processed by worker")
	log.Println("   Start worker: go run cmd/worker/main.go")
	log.Println("   Monitor tasks: http://localhost:8085 (Asynqmon)")

	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

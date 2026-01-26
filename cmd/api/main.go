package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/go-asynq-loadtest/internal/handler"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
	"github.com/lppduy/go-asynq-loadtest/internal/service"
	"github.com/lppduy/go-asynq-loadtest/pkg/database"
)

func main() {
	log.Println("🚀 Starting Order Processing API...")

	// Database configuration (from environment or defaults)
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "admin"),
		Password: getEnv("DB_PASSWORD", "secret123"),
		DBName:   getEnv("DB_NAME", "taskqueue"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
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

	// Initialize layers (Dependency Injection)
	orderRepo := repository.NewGormOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

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
	port := ":8080"
	log.Printf("✅ API server running on http://localhost%s", port)
	log.Println("📚 Endpoints:")
	log.Println("   - POST   /api/v1/orders          (Create order)")
	log.Println("   - GET    /api/v1/orders          (List orders)")
	log.Println("   - GET    /api/v1/orders/:id      (Get order)")
	log.Println("   - GET    /api/v1/orders/:id/status (Get status)")
	log.Println("   - POST   /api/v1/orders/:id/cancel (Cancel order)")
	log.Println("")
	log.Println("💡 Try: curl http://localhost:8080/health")

	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

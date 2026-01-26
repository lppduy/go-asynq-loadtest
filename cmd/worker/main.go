package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/lppduy/go-asynq-loadtest/internal/config"
	"github.com/lppduy/go-asynq-loadtest/internal/tasks"
)

func main() {
	log.Println("🔧 Starting Asynq Worker...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create Redis connection options
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create Asynq server with queue configuration
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Number of concurrent workers
			Concurrency: cfg.Worker.Concurrency,

			// Queue priority (higher number = higher priority)
			Queues: map[string]int{
				"critical": 6, // Highest priority (payment processing)
				"high":     4, // High priority (inventory updates)
				"default":  2, // Default priority (emails, invoices)
				"low":      1, // Low priority (analytics, notifications)
			},

			// Error handling
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx any, task *asynq.Task, err error) {
				log.Printf("❌ [Error] Task %s failed: %v", task.Type(), err)
			}),

			// Retry configuration
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,

			// Logger
			Logger: log.New(os.Stdout, "[asynq] ", log.LstdFlags),
		},
	)

	// Create task multiplexer (router)
	mux := asynq.NewServeMux()

	// Register task handlers
	// Critical queue
	mux.HandleFunc(tasks.TypePaymentProcess, tasks.HandlePaymentProcessTask)

	// High queue
	mux.HandleFunc(tasks.TypeInventoryUpdate, tasks.HandleInventoryUpdateTask)

	// Default queue
	mux.HandleFunc(tasks.TypeEmailConfirmation, tasks.HandleEmailConfirmationTask)
	mux.HandleFunc(tasks.TypeInvoiceGenerate, tasks.HandleInvoiceGenerateTask)

	// Low queue
	mux.HandleFunc(tasks.TypeAnalyticsTrack, tasks.HandleAnalyticsTrackTask)
	mux.HandleFunc(tasks.TypeWarehouseNotify, tasks.HandleWarehouseNotifyTask)

	log.Println("✅ Worker registered task handlers:")
	log.Println("   💳 [Critical] payment:process")
	log.Println("   📦 [High]     inventory:update")
	log.Println("   📧 [Default]  email:confirmation")
	log.Println("   🧾 [Default]  invoice:generate")
	log.Println("   📊 [Low]      analytics:track")
	log.Println("   🏭 [Low]      warehouse:notify")
	log.Println("")
	log.Printf("⚙️  Worker concurrency: %d", cfg.Worker.Concurrency)
	log.Printf("🔴 Redis: %s", cfg.Redis.Addr)
	log.Println("")
	log.Println("🚀 Worker started! Waiting for tasks...")

	// Start worker in a goroutine
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Failed to run worker: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("🛑 Shutting down worker gracefully...")

	// Graceful shutdown
	srv.Shutdown()

	log.Println("✅ Worker stopped successfully")
}

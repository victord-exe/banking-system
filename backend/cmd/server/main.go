package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/account"
	"github.com/hlabs/banking-system/internal/auth"
	"github.com/hlabs/banking-system/internal/chat"
	"github.com/hlabs/banking-system/internal/config"
	"github.com/hlabs/banking-system/internal/database"
	"github.com/hlabs/banking-system/internal/routes"
	"github.com/hlabs/banking-system/internal/tigerbeetle"
	"github.com/hlabs/banking-system/internal/transaction"
)

const (
	// TigerBeetle connection retry configuration
	maxConnectionRetries = 10
	retryDelaySeconds    = 2
)

func main() {
	log.Println("🚀 Starting HLABS Banking System Backend...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Println("✅ Configuration loaded")

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Connect to PostgreSQL
	db, err := database.Connect(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Connect to TigerBeetle with retry logic
	log.Println("🔗 Connecting to TigerBeetle...")
	var tbClient *tigerbeetle.Client

	for attempt := 1; attempt <= maxConnectionRetries; attempt++ {
		log.Printf("   Attempt %d/%d to connect to TigerBeetle at %s", attempt, maxConnectionRetries, cfg.TigerBeetleAddress)

		tbClient, err = tigerbeetle.NewClient(cfg.TigerBeetleAddress)
		if err == nil {
			log.Println("✅ Successfully connected to TigerBeetle!")
			break
		}

		log.Printf("⚠️  Connection attempt %d failed: %v", attempt, err)

		if attempt < maxConnectionRetries {
			log.Printf("   Retrying in %d seconds...", retryDelaySeconds)
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
		} else {
			log.Fatalf("❌ Failed to connect to TigerBeetle after %d attempts: %v", maxConnectionRetries, err)
		}
	}
	defer tbClient.Close()

	// Seed database with test users (if needed)
	if err := database.Seed(db, tbClient); err != nil {
		log.Printf("⚠️  Warning: Failed to seed database: %v", err)
		// Non-fatal: continue even if seeding fails
	}

	// Initialize services
	accountService := account.NewService(db, tbClient)
	transactionService := transaction.NewService(db, tbClient)
	chatService := chat.NewService(accountService, transactionService)

	// Initialize handlers
	authHandler := auth.NewHandler(db, tbClient, cfg.JWTSecret)
	accountHandler := account.NewHandler(accountService)
	transactionHandler := transaction.NewHandler(transactionService)
	chatHandler := chat.NewHandler(chatService)

	// Setup Gin router
	router := gin.Default()

	// Setup all routes
	routes.SetupRoutes(router, authHandler, accountHandler, transactionHandler, chatHandler, cfg.JWTSecret)

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("\n🛑 Shutting down server...")

		// Close database connection
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}

		os.Exit(0)
	}()

	// Start server
	serverAddr := ":" + cfg.ServerPort
	log.Printf("🌐 Server listening on http://localhost%s", serverAddr)
	log.Printf("📋 API Documentation: http://localhost%s/api", serverAddr)
	log.Printf("💚 Health Check: http://localhost%s/health", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hlabs/banking-system/internal/account"
	"github.com/hlabs/banking-system/internal/auth"
	"github.com/hlabs/banking-system/internal/chat"
	"github.com/hlabs/banking-system/internal/middleware"
	"github.com/hlabs/banking-system/internal/transaction"
)

// SetupRoutes configures all API routes for the application
func SetupRoutes(
	router *gin.Engine,
	authHandler *auth.Handler,
	accountHandler *account.Handler,
	transactionHandler *transaction.Handler,
	chatHandler *chat.Handler,
	jwtSecret string,
) {
	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "hlabs-banking-api",
		})
	})

	// API routes group
	api := router.Group("/api")
	{
		// ========================================
		// Public routes - Authentication
		// ========================================
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/logout", authHandler.Logout)
		}

		// ========================================
		// Protected routes - Accounts
		// ========================================
		accountRoutes := api.Group("/accounts")
		accountRoutes.Use(middleware.AuthMiddleware(jwtSecret))
		{
			accountRoutes.GET("/me", accountHandler.GetAccountInfo)
			accountRoutes.GET("/balance", accountHandler.GetBalance)
		}

		// ========================================
		// Protected routes - Transactions
		// ========================================
		transactionRoutes := api.Group("/transactions")
		transactionRoutes.Use(middleware.AuthMiddleware(jwtSecret))
		{
			transactionRoutes.POST("/deposit", transactionHandler.Deposit)
			transactionRoutes.POST("/withdraw", transactionHandler.Withdraw)
			transactionRoutes.POST("/transfer", transactionHandler.Transfer)
			transactionRoutes.GET("/history", transactionHandler.GetHistory)
		}

		// ========================================
		// Protected routes - AI Chat
		// ========================================
		chatRoutes := api.Group("/chat")
		chatRoutes.Use(middleware.AuthMiddleware(jwtSecret))
		{
			chatRoutes.POST("", chatHandler.ProcessMessage)
		}
	}
}

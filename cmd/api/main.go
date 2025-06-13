package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/auth/middleware"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/database"
	ahandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/auth/handlers"
	uhandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/handlers"
	whandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/writers/handlers"
	ohandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME environment variable not set")
	}
	client, err := database.ConnectMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer database.DisconnectMongoDB(client)

	r := gin.Default()

	// Initialize Handlers (you'll need to pass in services and database client)
	authHandler := ahandlers.NewAuthHandler(client, dbName)
	userHandler := uhandlers.NewUserHandler(client, dbName)
	writerHandler := whandlers.NewWriterHandler(client, dbName)
	orderHandler := ohandlers.NewOrderHandler(client, dbName)

	// Public Routes
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// User Routes
		protected.POST("/orders", userHandler.CreateOrder)
		protected.GET("/orders/me", userHandler.GetUserOrders)

		// Writer Routes (Admin protected)
		writers := protected.Group("/writers")
		writers.Use(middleware.AdminRoleMiddleware())
		{
			writers.POST("/", writerHandler.CreateWriter)
			writers.GET("/", writerHandler.ListWriters)
			writers.GET("/:id", writerHandler.GetWriterByID)
			writers.PUT("/:id", writerHandler.UpdateWriter)
			writers.DELETE("/:id", writerHandler.DeleteWriter)
		}

		// Admin Routes
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminRoleMiddleware())
		{
			admin.GET("/orders", orderHandler.ListOrders)
			admin.GET("/orders/submitted", orderHandler.ListSubmittedOrders)
			admin.PUT("/orders/:id/assign", orderHandler.AssignOrder)
		}

		// Writer Specific Routes
		writer := protected.Group("/writer")
		writer.Use(middleware.WriterRoleMiddleware())
		{
			writer.POST("/orders/:id/submit", orderHandler.SubmitOrder)
		}

		// Order Review Routes (User protected for approval/feedback)
		orderReview := protected.Group("/orders/:id/review")
		orderReview.Use(middleware.AuthMiddleware())
		{
			orderReview.PUT("/approve", orderHandler.ApproveOrder)
			orderReview.PUT("/feedback", orderHandler.ProvideFeedback)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Server started on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

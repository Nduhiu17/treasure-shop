package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	ahandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/auth/handlers"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/auth/middleware"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/database"
	ohandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/handlers"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	uhandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/handlers"
	userrolehandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/handlers"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	whandlers "github.com/nduhiu17/treasure-shop/cmd/api/internal/writers/handlers"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
)

func registerRoleRoutes(r *gin.Engine, db *mongo.Database) {
	roleService := userservices.NewRoleService(db)
	roleHandler := userrolehandlers.NewRoleHandler(roleService)
	userRoleService := userservices.NewUserRoleService(db)
	userRoleHandler := userrolehandlers.NewUserRoleHandler(userRoleService)

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/roles", roleHandler.Create)
		admin.GET("/roles", roleHandler.List)
		admin.GET("/roles/:id", roleHandler.GetByID)
		admin.PUT("/roles/:id", roleHandler.Update)
		admin.DELETE("/roles/:id", roleHandler.Delete)

		admin.POST("/user-roles", userRoleHandler.Create)
		admin.GET("/user-roles", userRoleHandler.List)
		admin.GET("/user-roles/user/:user_id", userRoleHandler.GetByUserID)
		admin.DELETE("/user-roles/:id", userRoleHandler.Delete)
	}
}

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

	db := client.Database(dbName)
	roleService := userservices.NewRoleService(db)
	userRoleService := userservices.NewUserRoleService(db)

	r := gin.Default()

	// Initialize Handlers (you'll need to pass in services and database client)
	authHandler := ahandlers.NewAuthHandler(client, dbName, userRoleService, roleService)
	userHandler := uhandlers.NewUserHandler(client, dbName)
	writerHandler := whandlers.NewWriterHandler(client, dbName)
	orderHandler := ohandlers.NewOrderHandler(client, dbName)

	// OrderType Service/Handler
	orderTypeCol := client.Database(dbName).Collection("order_types")
	orderTypeService := ohandlers.NewOrderTypeHandler(
		services.NewOrderTypeService(orderTypeCol),
	)

	// Public Routes
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Serve OpenAPI YAML directly
	r.StaticFile("/openapi.yaml", "./openapi.yaml")

	// Serve Swagger UI (using gin-swagger)
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

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

			// OrderType CRUD (admin only)
			admin.POST("/order-types", orderTypeService.Create)
			admin.GET("/order-types", orderTypeService.List)
			admin.GET("/order-types/:id", orderTypeService.GetByID)
			admin.PUT("/order-types/:id", orderTypeService.Update)
			admin.DELETE("/order-types/:id", orderTypeService.Delete)

			// List users by role (admin/super_admin only)
			admin.GET("/users", userHandler.ListUsersByRole)
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

	// Register role and user_role admin routes
	registerRoleRoutes(r, client.Database(dbName))

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

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetOrderSubmissions returns all submissions for a given order, sorted by creation date descending
func (h *OrderHandler) GetOrderSubmissions(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	oid, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	db := h.service.GetDB()
	submissionsCol := db.Collection("order_submissions")

	filter := bson.M{"order_id": oid}
	cursor, err := submissionsCol.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}
	defer cursor.Close(context.Background())

	var submissions []models.OrderSubmission
	if err := cursor.All(context.Background(), &submissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode submissions"})
		return
	}

	// Sort by SubmissionDate descending
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].SubmissionDate.After(submissions[j].SubmissionDate)
	})

	c.JSON(http.StatusOK, gin.H{"submissions": submissions})
}

type OrderHandler struct {
	service *services.OrderService
	db      *mongo.Database
}

func NewOrderHandler(client *mongo.Client, dbName string) *OrderHandler {
	return &OrderHandler{
		service: services.NewOrderService(client.Database(dbName)),
		db:      client.Database(dbName),
	}
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Pagination params
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Filtering params
	var userIDPtr *primitive.ObjectID
	var writerIDPtr *primitive.ObjectID
	var statusPtr *string
	if userID := c.Query("user_id"); userID != "" {
		if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
			userIDPtr = &oid
		}
	}
	if writerID := c.Query("writer_id"); writerID != "" {
		if oid, err := primitive.ObjectIDFromHex(writerID); err == nil {
			writerIDPtr = &oid
		}
	}
	if status := c.Query("status"); status != "" {
		statusPtr = &status
	}

	orders, total, err := h.service.GetOrdersFiltered(userIDPtr, writerIDPtr, statusPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}
	// Populate LevelName
	orderLevelService := services.NewOrderLevelService(h.db)
	orders = services.PopulateOrderLevelNames(orders, orderLevelService)
	// Populate OrderPagesName
	orderPagesService := services.NewOrderPagesService(h.db)
	orders = services.PopulateOrderPagesNames(orders, orderPagesService)
	// Populate OrderUrgencyName
	orderUrgencyService := services.NewOrderUrgencyService(h.db)
	orders = services.PopulateOrderUrgencyNames(orders, orderUrgencyService)
	// Populate OrderStyleName
	orderStyleService := services.NewOrderStyleService(h.db)
	orders = services.PopulateOrderStyleNames(orders, orderStyleService)
	// Populate OrderLanguageName
	orderLanguageService := services.NewOrderLanguageService(h.db)
	orders = services.PopulateOrderLanguageNames(orders, orderLanguageService)
	// Populate WriterName
	userService := userservices.NewUserService(h.db)
	orders = services.PopulateWriterNames(orders, userService)

	// Return orders list without writer_submissions field
	c.JSON(http.StatusOK, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *OrderHandler) ListSubmittedOrders(c *gin.Context) {
	orders, err := h.service.GetOrdersByStatus("submitted_for_review")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list submitted orders"})
		return
	}
	orderUrgencyService := services.NewOrderUrgencyService(h.db)
	orders = services.PopulateOrderUrgencyNames(orders, orderUrgencyService)
	// Populate OrderStyleName
	orderStyleService := services.NewOrderStyleService(h.db)
	orders = services.PopulateOrderStyleNames(orders, orderStyleService)
	// Populate OrderLanguageName
	orderLanguageService := services.NewOrderLanguageService(h.db)
	orders = services.PopulateOrderLanguageNames(orders, orderLanguageService)
	// Populate WriterName
	userService := userservices.NewUserService(h.db)
	orders = services.PopulateWriterNames(orders, userService)
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) AssignOrder(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var assignRequest struct {
		WriterID string `json:"writer_id"`
	}
	if err := c.ShouldBindJSON(&assignRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	writerOID, err := primitive.ObjectIDFromHex(assignRequest.WriterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format"})
		return
	}

	if err := h.service.AssignOrder(orderOID, writerOID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order assigned successfully"})
}

func (h *OrderHandler) SubmitOrder(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var submitRequest struct {
		SubmissionFile        string `json:"active_submission_file" binding:"required"`
		SubmissionDescription string `json:"active_submission_writer_note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&submitRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	writerIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Writer ID not found"})
		return
	}
	writerOID, err := primitive.ObjectIDFromHex(writerIDInterface.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid writer ID format"})
		return
	}

	if err := h.service.SubmitOrder(orderOID, writerOID, submitRequest.SubmissionFile, submitRequest.SubmissionDescription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order submitted for review"})
}

func (h *OrderHandler) ApproveOrder(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
		return
	}
	userOID, err := primitive.ObjectIDFromHex(userIDInterface.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if err := h.service.ApproveOrder(orderOID, userOID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order approved"})
}

func (h *OrderHandler) ProvideFeedback(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
		return
	}
	userOID, err := primitive.ObjectIDFromHex(userIDInterface.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var feedbackRequest struct {
		Feedback     string `json:"feedback"`
		FeedbackFile string `json:"feed_back_file"`
	}
	if err := c.ShouldBindJSON(&feedbackRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ProvideFeedback(orderOID, userOID, feedbackRequest.Feedback, feedbackRequest.FeedbackFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provide feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback submitted"})
}

// WriterAssignmentResponseRequest is the request body for writer assignment response
// Accept: true to accept, false to decline
type WriterAssignmentResponseRequest struct {
	Accept *bool `json:"accept" binding:"required"`
}

// WriterAcceptAssignment allows a writer to accept or decline an order assignment
func (h *OrderHandler) WriterAcceptAssignment(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID in context is not a string"})
		return
	}
	writerOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format"})
		return
	}
	var req WriterAssignmentResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.WriterAssignmentResponse(orderOID, writerOID, *req.Accept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment status", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assignment response recorded"})
}

// GetOrdersByWriter returns all orders assigned to a specific writer
func (h *OrderHandler) GetOrdersByWriter(c *gin.Context) {
	writerID := c.Param("writer_id")
	if writerID == "" {
		// Try to get from query param as fallback for misrouted requests
		writerID = c.Query("writer_id")
	}
	if len(writerID) != 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format. Must be a 24-character hex string ObjectID."})
		return
	}
	writerOID, err := primitive.ObjectIDFromHex(writerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format. Must be a valid hex ObjectID."})
		return
	}
	// Optional status filter
	var statusPtr *string
	status := c.Query("status")
	if status != "" {
		statusPtr = &status
	}
	// Optional pagination
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	orders, total, err := h.service.GetOrdersFiltered(nil, &writerOID, statusPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders for writer"})
		return
	}
	// Populate LevelName
	orderLevelService := services.NewOrderLevelService(h.db)
	orders = services.PopulateOrderLevelNames(orders, orderLevelService)
	// Populate OrderPagesName
	orderPagesService := services.NewOrderPagesService(h.db)
	orders = services.PopulateOrderPagesNames(orders, orderPagesService)
	// Populate OrderUrgencyName
	orderUrgencyService := services.NewOrderUrgencyService(h.db)
	orders = services.PopulateOrderUrgencyNames(orders, orderUrgencyService)
	// Populate OrderStyleName
	orderStyleService := services.NewOrderStyleService(h.db)
	orders = services.PopulateOrderStyleNames(orders, orderStyleService)
	// Populate OrderLanguageName
	orderLanguageService := services.NewOrderLanguageService(h.db)
	orders = services.PopulateOrderLanguageNames(orders, orderLanguageService)
	// Populate WriterName
	userService := userservices.NewUserService(h.db)
	orders = services.PopulateWriterNames(orders, userService)
	// Return orders list without writer_submissions field
	c.JSON(http.StatusOK, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate required fields
	if order.OrderTypeID.IsZero() || order.OrderLevelID.IsZero() || order.OrderUrgencyID.IsZero() || order.OrderPagesID.IsZero() || order.Price == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required order fields: order_type_id, order_level_id, order_urgency_id, order_pages_id, price"})
		return
	}
	insertedID, err := h.service.CreateOrder(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order", "details": err.Error()})
		return
	}
	order.ID = insertedID
	c.JSON(http.StatusCreated, order)
}

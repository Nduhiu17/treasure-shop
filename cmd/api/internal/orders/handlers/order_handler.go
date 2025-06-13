package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(client *mongo.Client,dbName string) *OrderHandler {
	return &OrderHandler{
		service: services.NewOrderService(client.Database(dbName)),
	}
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) ListSubmittedOrders(c *gin.Context) {
	orders, err := h.service.GetOrdersByStatus("submitted_for_review")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list submitted orders"})
		return
	}
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
		Content string `json:"content"`
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

	if err := h.service.SubmitOrder(orderOID, writerOID, submitRequest.Content); err != nil {
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
		Feedback string `json:"feedback"`
	}
	if err := c.ShouldBindJSON(&feedbackRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ProvideFeedback(orderOID, userOID, feedbackRequest.Feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provide feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback submitted"})
}

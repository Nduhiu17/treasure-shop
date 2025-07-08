package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedbackHandler struct {
	DB *mongo.Database
}

func NewFeedbackHandler(db *mongo.Database) *FeedbackHandler {
	return &FeedbackHandler{DB: db}
}

// GetOrderFeedbacks returns all feedbacks for a given order, sorted by created_at descending
func (h *FeedbackHandler) GetOrderFeedbacks(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	coll := h.DB.Collection("order_feedbacks")
	cursor, err := coll.Find(context.Background(), bson.M{"order_id": orderOID},
		&options.FindOptions{Sort: bson.D{{Key: "created_at", Value: -1}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedbacks"})
		return
	}
	defer cursor.Close(context.Background())
	var feedbacks []models.OrderFeedback
	if err := cursor.All(context.Background(), &feedbacks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode feedbacks"})
		return
	}
	if feedbacks == nil {
		feedbacks = []models.OrderFeedback{}
	}
	c.JSON(http.StatusOK, gin.H{"feedbacks": feedbacks})
}

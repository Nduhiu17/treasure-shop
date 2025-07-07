package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	orderservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderDetailsHandler struct {
	service *orderservices.OrderService
	db      *mongo.Database
}

func NewOrderDetailsHandler(client *mongo.Client, dbName string) *OrderDetailsHandler {
	return &OrderDetailsHandler{
		service: orderservices.NewOrderService(client.Database(dbName)),
		db:      client.Database(dbName),
	}
}

// GetOrderDetails returns order details with writer_submissions for any authenticated user
func (h *OrderDetailsHandler) GetOrderDetails(c *gin.Context) {
	orderID := c.Param("id")
	orderOID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}
	orderPtr, err := h.service.GetOrderByID(orderOID)
	if err != nil || orderPtr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	order := *orderPtr // use value, not pointer

	// Populate related fields
	orderLevelService := orderservices.NewOrderLevelService(h.db)
	order = orderservices.PopulateOrderLevelNames([]models.Order{order}, orderLevelService)[0]
	orderPagesService := orderservices.NewOrderPagesService(h.db)
	order = orderservices.PopulateOrderPagesNames([]models.Order{order}, orderPagesService)[0]
	orderUrgencyService := orderservices.NewOrderUrgencyService(h.db)
	order = orderservices.PopulateOrderUrgencyNames([]models.Order{order}, orderUrgencyService)[0]
	orderStyleService := orderservices.NewOrderStyleService(h.db)
	order = orderservices.PopulateOrderStyleNames([]models.Order{order}, orderStyleService)[0]
	orderLanguageService := orderservices.NewOrderLanguageService(h.db)
	order = orderservices.PopulateOrderLanguageNames([]models.Order{order}, orderLanguageService)[0]
	userService := userservices.NewUserService(h.db)
	order = orderservices.PopulateWriterNames([]models.Order{order}, userService)[0]

	// Get submissions (descending by submission_date)
	orderSubmissionColl := h.db.Collection("order_submissions")
	findOpts := &options.FindOptions{Sort: bson.D{{Key: "submission_date", Value: -1}}}
	cursor, err := orderSubmissionColl.Find(c, bson.M{"order_id": order.ID}, findOpts)
	var subs []models.OrderSubmission
	if err == nil {
		_ = cursor.All(c, &subs)
	} else {
		subs = []models.OrderSubmission{}
	}
	orderMap := gin.H{}
	orderBytes, _ := json.Marshal(order)
	_ = json.Unmarshal(orderBytes, &orderMap)
	orderMap["writer_submissions"] = subs

	c.JSON(http.StatusOK, orderMap)
}

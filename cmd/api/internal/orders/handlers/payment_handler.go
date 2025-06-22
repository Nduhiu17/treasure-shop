package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	orderservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/payments/services"
	"go.mongodb.org/mongo-driver/mongo"
)

// PaymentRequest represents the expected payload for payment
// method: "paypal" or "mastercard"
// paymentInfo: map with required fields for the gateway
// order_id: the order to pay for

type PaymentRequest struct {
	OrderID     string                 `json:"order_id" binding:"required"`
	Method      string                 `json:"method" binding:"required,oneof=paypal mastercard"`
	PaymentInfo map[string]interface{} `json:"payment_info" binding:"required"`
}

// PayForOrderHandler allows a user to pay for an order using PayPal or Mastercard
func PayForOrderHandler(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentService := services.NewPaymentService()
	// Get DB from Gin context (set in main.go)
	dbIface, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found in context"})
		return
	}
	db, ok := dbIface.(*mongo.Database)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection in context"})
		return
	}
	orderService := orderservices.NewOrderService(db)

	if req.Method == "paypal" {
		amount, _ := req.PaymentInfo["amount"].(float64)
		currency, _ := req.PaymentInfo["currency"].(string)
		returnURL, _ := req.PaymentInfo["return_url"].(string)
		cancelURL, _ := req.PaymentInfo["cancel_url"].(string)
		orderID, err := paymentService.CreatePayPalOrder(amount, currency, returnURL, cancelURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PayPal order", "details": err.Error()})
			return
		}
		// Mark order as paid
		if err := orderService.MarkOrderPaid(req.OrderID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment succeeded but failed to update order status", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"paypal_order_id": orderID})
		return
	}

	if req.Method == "paypal_capture" {
		paypalOrderID, _ := req.PaymentInfo["paypal_order_id"].(string)
		status, err := paymentService.CapturePayPalOrder(paypalOrderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to capture PayPal order", "details": err.Error()})
			return
		}
		// Mark order as paid
		if err := orderService.MarkOrderPaid(req.OrderID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment succeeded but failed to update order status", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
		return
	}

	// Mastercard and other methods can be handled here

	success, err := paymentService.ProcessPayment(req.OrderID, req.PaymentInfo)
	if err != nil || !success {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment failed", "details": err})
		return
	}
	// Mark order as paid
	if err := orderService.MarkOrderPaid(req.OrderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment succeeded but failed to update order status", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payment successful"})
}

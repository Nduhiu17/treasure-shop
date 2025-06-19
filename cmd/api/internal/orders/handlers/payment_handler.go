package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/payments/services"
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
	success, err := paymentService.ProcessPayment(req.OrderID, req.PaymentInfo)
	if err != nil || !success {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment failed", "details": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment successful"})
}

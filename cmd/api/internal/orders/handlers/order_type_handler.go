package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderTypeHandler struct {
	Service *services.OrderTypeService
}

func NewOrderTypeHandler(service *services.OrderTypeService) *OrderTypeHandler {
	return &OrderTypeHandler{Service: service}
}

// Middleware to check admin or super admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" && role != "super_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admins only"})
			return
		}
		c.Next()
	}
}

func (h *OrderTypeHandler) Create(c *gin.Context) {
	var orderType models.OrderType
	if err := c.ShouldBindJSON(&orderType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID is not a string"})
		return
	}
	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	orderType.CreatedBy = objID
	if err := h.Service.Create(context.Background(), &orderType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, orderType)
}

func (h *OrderTypeHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	orderType, err := h.Service.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OrderType not found"})
		return
	}
	c.JSON(http.StatusOK, orderType)
}

func (h *OrderTypeHandler) List(c *gin.Context) {
	orderTypes, err := h.Service.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orderTypes)
}

func (h *OrderTypeHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var update map[string]interface{}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.Update(context.Background(), id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OrderType updated"})
}

func (h *OrderTypeHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.Service.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OrderType deleted"})
}

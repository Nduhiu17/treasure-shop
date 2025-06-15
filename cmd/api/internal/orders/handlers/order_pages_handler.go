package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderPagesHandler struct {
	service *services.OrderPagesService
}

func NewOrderPagesHandler(service *services.OrderPagesService) *OrderPagesHandler {
	return &OrderPagesHandler{service: service}
}

func (h *OrderPagesHandler) Create(c *gin.Context) {
	var pages models.OrderPages
	if err := c.ShouldBindJSON(&pages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Create(&pages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pages)
}

func (h *OrderPagesHandler) List(c *gin.Context) {
	pages, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pages)
}

func (h *OrderPagesHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	pages, err := h.service.GetByID(oid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order pages not found"})
		return
	}
	c.JSON(http.StatusOK, pages)
}

func (h *OrderPagesHandler) Update(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var update bson.M
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Update(oid, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order pages updated"})
}

func (h *OrderPagesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.service.Delete(oid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order pages deleted"})
}

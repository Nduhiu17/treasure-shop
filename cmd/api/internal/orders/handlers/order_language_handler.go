package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderLanguageHandler struct {
	service *services.OrderLanguageService
}

func NewOrderLanguageHandler(service *services.OrderLanguageService) *OrderLanguageHandler {
	return &OrderLanguageHandler{service: service}
}

func (h *OrderLanguageHandler) Create(c *gin.Context) {
	var language models.OrderLanguage
	if err := c.ShouldBindJSON(&language); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Create(&language); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, language)
}

func (h *OrderLanguageHandler) List(c *gin.Context) {
	languages, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, languages)
}

func (h *OrderLanguageHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	language, err := h.service.GetByID(oid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order language not found"})
		return
	}
	c.JSON(http.StatusOK, language)
}

func (h *OrderLanguageHandler) Update(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "Order language updated"})
}

func (h *OrderLanguageHandler) Delete(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "Order language deleted"})
}

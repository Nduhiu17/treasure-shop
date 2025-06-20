package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WriterHandler struct {
	service *services.UserService
	db      *mongo.Database
}

func NewWriterHandler(client *mongo.Client, dbName string) *WriterHandler {
	db := client.Database(dbName)
	return &WriterHandler{
		service: services.NewUserService(db),
		db:      db,
	}
}

func (h *WriterHandler) CreateWriter(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	roleService := services.NewRoleService(h.db)
	userRoleService := services.NewUserRoleService(h.db)
	writerRole, err := roleService.GetByName("writer")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Writer role not found. Please create it first."})
		return
	}
	if err := h.service.CreateUser(&user, userRoleService, roleService, writerRole.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create writer account"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Writer account created successfully"})
}

func (h *WriterHandler) ListWriters(c *gin.Context) {
	userRoleService := services.NewUserRoleService(h.db)
	roleService := services.NewRoleService(h.db)
	writers, err := h.service.GetUsersByRole("writer", userRoleService, roleService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list writers"})
		return
	}
	c.JSON(http.StatusOK, writers)
}

func (h *WriterHandler) GetWriterByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format"})
		return
	}
	writer, err := h.service.GetUserByID(objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Writer not found"})
		return
	}
	c.JSON(http.StatusOK, writer)
}

func (h *WriterHandler) UpdateWriter(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format"})
		return
	}
	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser.ID = objID // Ensure ID is set for update

	if err := h.service.UpdateUser(&updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update writer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Writer updated successfully"})
}

func (h *WriterHandler) DeleteWriter(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid writer ID format"})
		return
	}
	if err := h.service.DeleteUser(objID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete writer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Writer deleted successfully"})
}

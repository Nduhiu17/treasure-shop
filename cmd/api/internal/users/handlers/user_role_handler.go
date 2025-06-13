package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRoleHandler struct {
	service *userservices.UserRoleService
}

func NewUserRoleHandler(service *userservices.UserRoleService) *UserRoleHandler {
	return &UserRoleHandler{service: service}
}

func (h *UserRoleHandler) Create(c *gin.Context) {
	var userRole models.UserRole
	if err := c.ShouldBindJSON(&userRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Create(&userRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userRole)
}

func (h *UserRoleHandler) List(c *gin.Context) {
	userRoles, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userRoles)
}

func (h *UserRoleHandler) GetByUserID(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	userRoles, err := h.service.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userRoles)
}

func (h *UserRoleHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "UserRole deleted"})
}

// Add endpoint to assign a role to a user (user can have multiple roles)
func (h *UserRoleHandler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}
	roleID, err := primitive.ObjectIDFromHex(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id format"})
		return
	}
	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := h.service.Create(userRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned to user successfully"})
}

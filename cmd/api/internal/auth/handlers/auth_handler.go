package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/auth/services"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	service         *services.AuthService
	userRoleService *userservices.UserRoleService
	roleService     *userservices.RoleService
}

func NewAuthHandler(client *mongo.Client, dbName string, userRoleService *userservices.UserRoleService, roleService *userservices.RoleService) *AuthHandler {
	return &AuthHandler{
		service:         services.NewAuthService(client.Database(dbName)),
		userRoleService: userRoleService,
		roleService:     roleService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Register(&user, h.userRoleService, h.roleService); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.service.Login(credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	userservices "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	orderService    *services.OrderService
	userService     *userservices.UserService
	userRoleService *userservices.UserRoleService
	roleService     *userservices.RoleService
}

func NewUserHandler(client *mongo.Client, dbName string) *UserHandler {
	db := client.Database(dbName)
	return &UserHandler{
		orderService:    services.NewOrderService(db),
		userService:     userservices.NewUserService(db),
		userRoleService: userservices.NewUserRoleService(db),
		roleService:     userservices.NewRoleService(db),
	}
}

func (h *UserHandler) CreateOrder(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
		return
	}
	userID := userIDInterface.(string)
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order.UserID = userOID
	// Ensure OrderTypeID is provided and valid
	if order.OrderTypeID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderTypeID is required"})
		return
	}
	order.Status = "pending_payment" // Initial status

	if err := h.orderService.CreateOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *UserHandler) GetUserOrders(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
		return
	}
	userID := userIDInterface.(string)
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(userOID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *UserHandler) ListUsersByRole(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role query parameter is required (admin, user, super_admin)"})
		return
	}
	users, err := h.userService.GetUsersByRole(role, h.userRoleService, h.roleService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users by role"})
		return
	}
	// Populate roles for each user
	for i := range users {
		userRoles, err := h.userRoleService.GetByUserID(users[i].ID)
		if err != nil {
			continue // skip if error
		}
		var roleNames []string
		for _, ur := range userRoles {
			roleObj, err := h.roleService.GetByID(ur.RoleID)
			if err == nil {
				roleNames = append(roleNames, roleObj.Name)
			}
		}
		users[i].Roles = roleNames
	}
	c.JSON(http.StatusOK, users)
}

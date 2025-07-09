package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/models"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/orders/services"
	usermodels "github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in token/context"})
		return
	}
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
		return
	}
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
	fmt.Printf("[DEBUG] Incoming order payload: %+v\n", order)
	// Validate is_high_priority is present in the request (required)
	if c.Request.Method == "POST" && c.FullPath() == "/api/orders" {
		if c.PostForm("is_high_priority") == "" && !order.IsHighPriority {
			// If not present in JSON, and not set to true, default to false
			order.IsHighPriority = false
		}
	}
	// Enforce that is_high_priority is present (even if false)
	if c.Request.Method == "POST" && c.FullPath() == "/api/orders" && (order.IsHighPriority != true && order.IsHighPriority != false) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_high_priority is required (true or false)"})
		return
	}
	order.UserID = userOID
	order.WriterID = nil // Set WriterID to nil so it is omitted or null in JSON if not assigned
	// Ensure OrderTypeID is provided and valid
	if order.OrderTypeID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderTypeID is required"})
		return
	}
	// Ensure OrderUrgencyID is provided and valid
	if order.OrderUrgencyID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderUrgencyID is required"})
		return
	}
	// Ensure OrderStyleID is provided and valid
	if order.OrderStyleID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OrderStyleID is required"})
		return
	}
	// Set default values for new boolean fields
	order.TopWriter = false
	order.PlagarismReport = false
	order.OnePageSummary = false
	order.ExtraQualityCheck = false
	order.InitialDraft = false
	order.SmsUpdate = false
	order.FullTextCopySources = false
	order.SamePaperFromAnotherWriter = false
	order.Status = "pending_payment" // Initial status
	order.NoOfSources = "0"

	insertedID, err := h.orderService.CreateOrder(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order", "details": err.Error()})
		return
	}
	order.ID = insertedID
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

	// Pagination params
	page := 1
	pageSize := 8
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 8
	}
	// Status filter
	var statusPtr *string
	if status := c.Query("status"); status != "" {
		statusPtr = &status
	}
	// Use GetOrdersFiltered for filtering and pagination
	orders, total, err := h.orderService.GetOrdersFiltered(&userOID, nil, statusPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user orders"})
		return
	}
	// Enrich order details with related names
	orderLevelService := services.NewOrderLevelService(h.orderService.GetDB())
	orders = services.PopulateOrderLevelNames(orders, orderLevelService)
	orderPagesService := services.NewOrderPagesService(h.orderService.GetDB())
	orders = services.PopulateOrderPagesNames(orders, orderPagesService)
	orderUrgencyService := services.NewOrderUrgencyService(h.orderService.GetDB())
	orders = services.PopulateOrderUrgencyNames(orders, orderUrgencyService)
	orderStyleService := services.NewOrderStyleService(h.orderService.GetDB())
	orders = services.PopulateOrderStyleNames(orders, orderStyleService)
	orderLanguageService := services.NewOrderLanguageService(h.orderService.GetDB())
	orders = services.PopulateOrderLanguageNames(orders, orderLanguageService)
	userService := userservices.NewUserService(h.orderService.GetDB())
	orders = services.PopulateWriterNames(orders, userService)

	// Return orders list without writer_submissions field
	c.JSON(http.StatusOK, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *UserHandler) ListUsersByRole(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role query parameter is required (admin, user, super_admin)"})
		return
	}
	// Pagination params
	page := 1
	pageSize := 8
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 8
	}
	users, total, err := h.userService.GetUsersByRolePaginated(role, h.userRoleService, h.roleService, page, pageSize)
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
	if users == nil {
		users = []usermodels.User{}
	}
	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCurrentUser returns the details of the currently logged-in user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in token/context"})
		return
	}
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
		return
	}
	userOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	user, err := h.userService.GetUserByID(userOID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Populate roles
	userRoles, _ := h.userRoleService.GetByUserID(userOID)
	var roleNames []string
	for _, ur := range userRoles {
		roleObj, err := h.roleService.GetByID(ur.RoleID)
		if err == nil {
			roleNames = append(roleNames, roleObj.Name)
		}
	}
	user.Roles = roleNames
	user.Password = "" // never return password
	c.JSON(http.StatusOK, user)
}

// UpdateUser updates user details (admin/super_admin can update any user, others only their own)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" || len(userIDParam) != 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID parameter"})
		return
	}
	userOID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	// Get logged-in user info
	loggedInUserID, _ := c.Get("userID")
	loggedInUserOID, _ := primitive.ObjectIDFromHex(loggedInUserID.(string))
	// Get logged-in user roles
	userRoles, _ := h.userRoleService.GetByUserID(loggedInUserOID)
	var isAdminOrSuper bool
	for _, ur := range userRoles {
		roleObj, err := h.roleService.GetByID(ur.RoleID)
		if err == nil && (roleObj.Name == "admin" || roleObj.Name == "super_admin") {
			isAdminOrSuper = true
			break
		}
	}
	// Only admin/super_admin can update other users
	if !isAdminOrSuper && loggedInUserOID != userOID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own details"})
		return
	}
	// Fetch the user to update
	user, err := h.userService.GetUserByID(userOID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Bind update fields (ignore password and user_number)
	var updateData struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Tier      string `json:"tier"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if updateData.Email != "" {
		user.Email = updateData.Email
	}
	if updateData.Username != "" {
		user.Username = updateData.Username
	}
	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}
	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}
	if updateData.Tier != "" {
		user.Tier = updateData.Tier
	}
	// Do not update password or user_number
	err = h.userService.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

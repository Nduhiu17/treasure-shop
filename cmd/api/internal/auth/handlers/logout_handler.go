package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles user logout by instructing the client to delete the JWT
func LogoutHandler(c *gin.Context) {
	// For stateless JWT, logout is handled on the client by deleting the token.
	// Optionally, you can blacklist the token here if you implement token blacklisting.
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully. Please delete your token on the client."})
}

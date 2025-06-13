package handlers

import (
	"github.com/gin-gonic/gin"
)

// SwaggerUIHandler serves the Swagger UI and openapi.yaml
func SwaggerUIHandler(openapiPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File(openapiPath)
	}
}

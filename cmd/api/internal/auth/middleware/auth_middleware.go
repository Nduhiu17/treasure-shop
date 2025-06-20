package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/nduhiu17/treasure-shop/cmd/api/internal/users/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthMiddleware() gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Replace with a strong secret in .env
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["sub"].(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token claims"})
				return
			}
			fmt.Println("User ID from token:", userID)
			c.Set("userID", userID)

			rolesRaw := claims["roles"]
			var roles []interface{}
			switch v := rolesRaw.(type) {
			case []interface{}:
				roles = v
			case nil:
				roles = []interface{}{}
			case string:
				roles = []interface{}{v}
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid roles in token claims"})
				return
			}
			fmt.Println("Roles from token claims:", roles)
			c.Set("roles", roles)

			// Optionally fetch user details from the database here if needed for every request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			client, _ := c.Get("mongoClient") // You might need to set this up in your main function
			if client != nil {
				db := client.(*mongo.Client).Database("os.Getenv(\"DB_NAME\")") // Replace with your DB name
				var user models.User
				objID, _ := primitive.ObjectIDFromHex(userID)
				err := db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
				if err == nil {
					c.Set("user", user)
				}
			}

			// Set user_number in context if available in claims or user struct
			if userNumber, ok := claims["user_number"].(string); ok {
				c.Set("user_number", userNumber)
			} else if user, exists := c.Get("user"); exists {
				// Try to extract user_number from user struct (map or struct)
				switch u := user.(type) {
				case models.User:
					c.Set("user_number", u.UserNumber)
				case map[string]interface{}:
					if un, ok := u["user_number"].(string); ok {
						c.Set("user_number", un)
					}
				}
			}

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}

func AdminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, ok := c.Get("roles")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		for _, role := range roles.([]interface{}) {
			if role == "super_admin" || role == "admin" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
	}
}

func WriterRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, ok := c.Get("roles")
		fmt.Println("Checking writer role middleware", ok, roles)

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		for _, role := range roles.([]interface{}) {
			if role == "writer" || role == "admin" || role == "super_admin" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
	}
}

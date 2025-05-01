package delivery

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/usecase"
)

type AuthHandler struct {
	authUC usecase.AuthUseCase
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false})
		return
	}

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.authUC.SecretKey), nil
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

func extractToken(c *gin.Context) string {
	// 1. Проверяем Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString != "" {
		return strings.Replace(tokenString, "Bearer ", "", 1)
	}

	// 2. Проверяем cookie
	tokenString, _ = c.Cookie("access_token")
	if tokenString != "" {
		return tokenString
	}

	// 3. Проверяем query parameter
	tokenString = c.Query("token")
	return tokenString
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "missing_auth_header",
			})
			return
		}

		// Extract the token (handle Bearer prefix)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
				"code":  "invalid_token_format",
			})
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Auth.SecretKey), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": err.Error(),
				"code":    "invalid_token",
			})
			return
		}

		// Validate claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Handle user_id (it might be float64 or int)
			userID, ok := claims["user_id"].(float64)
			role, _ := claims["role"].(string)
			c.Set("user_role", role)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid user_id in token",
					"code":  "invalid_user_id",
				})
				return
			}

			username, ok := claims["username"].(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid username in token",
					"code":  "invalid_username",
				})
				return
			}

			// Set values in context
			c.Set("user_id", int(userID))
			c.Set("username", username)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "invalid_claims",
			})
		}
	}
}

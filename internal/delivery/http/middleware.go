package delivery

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/config"
)

type authUseCase interface {
	SecretKey() []byte
	GenerateToken(userID int, username string) (string, error)
	ParseToken(tokenString string) (int64, string, error)
}

// Изменяем AuthHandler для использования локального интерфейса
type AuthHandler struct {
	authUC authUseCase
}

func NewAuthHandler(authUC authUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// ValidateToken godoc
// @Summary Validate token
// @Description Validate JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 400 {object} docs.Error
// @Failure 401 {object} docs.Error
// @Router /auth/validate [get]

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	tokenString := extractToken(c)
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "token not provided"})
		return
	}

	valid, err := h.validateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (h *AuthHandler) validateJWT(tokenString string) (bool, error) {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Вызываем функцию SecretKey() чтобы получить []byte
		return h.authUC.SecretKey(), nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

func extractToken(c *gin.Context) string {
	// Check Authorization header first
	if token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "); token != "" {
		return token
	}

	// Then check cookie
	if token, _ := c.Cookie("access_token"); token != "" {
		return token
	}

	// Finally check query parameter
	return c.Query("token")
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		tokenString := extractToken(c)
		if tokenString == "" {
			abortWithAuthError(c, "Authorization token required", "missing_token")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Auth.SecretKey), nil
		})

		if err != nil {
			abortWithAuthError(c, "Invalid token", "invalid_token", "details", err.Error())
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, username, role, err := extractClaims(claims)
			if err != nil {
				abortWithAuthError(c, err.Error(), "invalid_claims")
				return
			}

			c.Set("user_id", userID)
			c.Set("username", username)
			c.Set("user_role", role)
			c.Next()
		} else {
			abortWithAuthError(c, "Invalid token claims", "invalid_claims")
		}
	}
}

func extractClaims(claims jwt.MapClaims) (int, string, string, error) {
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", "", fmt.Errorf("invalid user_id in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return 0, "", "", fmt.Errorf("invalid username in token")
	}

	role, _ := claims["role"].(string) // role is optional

	return int(userID), username, role, nil
}

func abortWithAuthError(c *gin.Context, errorMsg string, errorCode string, extra ...interface{}) {
	response := gin.H{
		"error": errorMsg,
		"code":  errorCode,
	}

	// Add extra fields if provided
	for i := 0; i < len(extra); i += 2 {
		if i+1 < len(extra) {
			key, ok := extra[i].(string)
			if ok {
				response[key] = extra[i+1]
			}
		}
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, response)
}

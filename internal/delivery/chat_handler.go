package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
)

type ChatHandler struct {
	chatUC usecase.ChatUseCase
}

func NewChatHandler(chatUC usecase.ChatUseCase) *ChatHandler {
	return &ChatHandler{chatUC: chatUC}
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	// Get user info from context with proper type assertions
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
			"code":  "missing_user_context",
		})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Username not found in context",
			"code":  "missing_username_context",
		})
		return
	}

	// Parse request
	var request struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "invalid_request_format",
		})
		return
	}

	// Create message with proper type assertions
	message := &entity.ChatMessage{
		UserID: userID.(int),
		Author: username.(string),
		Text:   request.Text,
	}

	// Save to database
	if err := h.chatUC.SendMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save message",
			"details": err.Error(),
			"code":    "database_error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         message.ID,
		"user_id":    message.UserID,
		"author":     message.Author,
		"text":       message.Text,
		"created_at": message.CreatedAt,
	})
}
func (h *ChatHandler) GetMessages(c *gin.Context) {
	messages, err := h.chatUC.GetMessages(c.Request.Context(), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

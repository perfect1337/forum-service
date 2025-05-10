// internal/delivery/http/chat_handler.go
package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatHandler struct {
	chatUC usecase.ChatUseCaseInterface
}

func NewChatHandler(chatUC usecase.ChatUseCaseInterface) *ChatHandler {
	return &ChatHandler{chatUC: chatUC}
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}
	h.chatUC.HandleWebSocket(conn)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
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

	message := &entity.ChatMessage{
		UserID: userID.(int),
		Author: username.(string),
		Text:   request.Text,
	}

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
	messages, err := h.chatUC.GetMessages(c.Request.Context(), 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

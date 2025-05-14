package delivery

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
)

type CommentHandler struct {
	commentUC usecase.CommentUseCaseInterface
}

func NewCommentHandler(commentUC usecase.CommentUseCaseInterface) *CommentHandler {
	return &CommentHandler{commentUC: commentUC}
}

// CreateComment godoc
// @Summary Create a comment
// @Description Create a new comment for a specific post. Requires Bearer token authentication.
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "Post ID"
// @Param comment body entity.Comment true "Comment object" SchemaExample({"content":"This is a comment"})
// @Success 201 {object} entity.Comment
// @Failure 400 {object} docs.Error "Invalid request format"
// @Failure 401 {object} docs.Error "Missing or invalid authentication token"
// @Failure 500 {object} docs.Error "Server error"
// @Router /posts/{id}/comments [post]

func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var comment entity.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Убедимся, что post_id берется из URL, а не из тела запроса
	comment.PostID = postID
	comment.UserID = userID.(int)

	if err := h.commentUC.CreateComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetComments godoc
// @Summary Get comments for a post
// @Description Retrieve all comments for a specific post
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {array} entity.Comment
// @Failure 400 {object} docs.Error
// @Failure 500 {object} docs.Error
// @Router /posts/{id}/comments [get]

func (h *CommentHandler) GetComments(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id")) // Преобразуем строку в int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// DeleteComment godoc
// @Summary Delete comment
// @Description Delete a specific comment
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param comment_id path int true "Comment ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} docs.Error
// @Failure 401 {object} docs.Error
// @Failure 404 {object} docs.Error
// @Failure 500 {object} docs.Error
// @Router /posts/{id}/comments/{comment_id} [delete]

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Преобразуем userID в int
	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	if err := h.commentUC.DeleteComment(
		c.Request.Context(),
		commentID,
		userIDInt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found or not authorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}

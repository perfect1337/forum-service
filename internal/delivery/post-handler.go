package delivery

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
)

type PostHandler struct {
	postUC    usecase.PostUseCase
	commentUC usecase.CommentUseCase
}

func NewPostHandler(postUC usecase.PostUseCase, commentUC usecase.CommentUseCase) *PostHandler {
	return &PostHandler{
		postUC:    postUC,
		commentUC: commentUC,
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var post entity.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Author = fmt.Sprintf("%v", userID)

	if err := h.postUC.CreatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := h.postUC.GetPostByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"post":     post,
		"comments": comments,
	}

	c.JSON(http.StatusOK, response)
}

func (h *PostHandler) GetAllPosts(c *gin.Context) {
	includeComments := c.Query("includeComments") == "true"

	posts, err := h.postUC.GetAllPosts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if includeComments {
		for i := range posts {
			comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), posts[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			posts[i].Comments = comments
		}
	}

	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) DeletePost(c *gin.Context) {

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	// Для не-админов проверяем авторство

	if err := h.postUC.DeletePost(c.Request.Context(), postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

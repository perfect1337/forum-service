package delivery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
)

type PostHandler struct {
	postUC    usecase.PostUseCase
	commentUC usecase.CommentUseCaseInterface
	userUC    usecase.UserUseCaseInterface
}

func NewPostHandler(
	postUC usecase.PostUseCase,
	commentUC usecase.CommentUseCaseInterface,
	userUC usecase.UserUseCaseInterface,
) *PostHandler {
	return &PostHandler{
		postUC:    postUC,
		commentUC: commentUC,
		userUC:    userUC,
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

	post.UserID = userID.(int)

	// Получаем имя автора
	user, err := h.userUC.GetUserByID(c.Request.Context(), post.UserID)
	if err == nil {
		post.Author = user.Username
	}

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

	// Получаем имя автора
	user, err := h.userUC.GetUserByID(c.Request.Context(), post.UserID)
	if err == nil {
		post.Author = user.Username
	}

	comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем имена авторов комментариев
	for i := range comments {
		user, err := h.userUC.GetUserByID(c.Request.Context(), comments[i].UserID)
		if err == nil {
			comments[i].Author = user.Username
		}
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

	// Получаем имена авторов для постов
	for i := range posts {
		user, err := h.userUC.GetUserByID(c.Request.Context(), posts[i].UserID)
		if err == nil {
			posts[i].Author = user.Username
		}

		if includeComments {
			comments, err := h.commentUC.GetCommentsByPostID(c.Request.Context(), posts[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Получаем имена авторов для комментариев
			for j := range comments {
				user, err := h.userUC.GetUserByID(c.Request.Context(), comments[j].UserID)
				if err == nil {
					comments[j].Author = user.Username
				}
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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.postUC.DeletePost(c.Request.Context(), postID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

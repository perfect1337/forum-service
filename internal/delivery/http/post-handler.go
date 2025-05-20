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

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new forum post. Requires Bearer token authentication.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <token>"
// @Param post body entity.Post true "Post object" SchemaExample({"title":"My Post","content":"Post content"})
// @Success 201 {object} entity.Post
// @Failure 400 {object} docs.Error "Invalid request format"
// @Failure 401 {object} docs.Error "Missing or invalid authentication token"
// @Failure 500 {object} docs.Error "Server error"
// @Router /posts [post]

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

// GetPostByID godoc
// @Summary Get post by ID
// @Description Retrieve a specific post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} entity.Post
// @Failure 400 {object} docs.Error
// @Failure 404 {object} docs.Error
// @Failure 500 {object} docs.Error
// @Router /posts/{id} [get]

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

// GetAllPosts godoc
// @Summary Get all posts
// @Description Retrieve a list of all forum posts
// @Tags posts
// @Accept json
// @Produce json
// @Param includeComments query boolean false "Include comments in response"
// @Success 200 {array} entity.Post
// @Failure 500 {object} docs.Error
// @Router /posts [get]

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

// DeletePost godoc
// @Summary Delete post
// @Description Delete a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} docs.Error
// @Failure 401 {object} docs.Error
// @Failure 500 {object} docs.Error
// @Router /posts/{id} [delete]

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

// UpdatePost godoc
// @Summary Update post
// @Description Update a specific post. Only the owner or admin can update.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param post body entity.Post true "Post object"
// @Success 200 {object} entity.Post
// @Failure 400 {object} docs.Error
// @Failure 401 {object} docs.Error
// @Failure 403 {object} docs.Error
// @Failure 404 {object} docs.Error
// @Failure 500 {object} docs.Error
// @Router /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
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

	var req entity.Post
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.postUC.UpdatePost(c.Request.Context(), postID, userID.(int), req.Title, req.Content)
	if err != nil {
		if err.Error() == "unauthorized: you can only update your own posts" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedPost, err := h.postUC.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedPost)
}

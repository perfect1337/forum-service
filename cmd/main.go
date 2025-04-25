package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/delivery"
	"github.com/perfect1337/forum-service/internal/repository"
	"github.com/perfect1337/forum-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Status(200)
	})

	router.Use(func(c *gin.Context) {
		log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL)
		log.Printf("Headers: %v", c.Request.Header)
		c.Next()
	})

	// Инициализация репозитория
	repo, err := repository.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	// Инициализация usecases
	postUC := usecase.NewPostUseCase(repo)
	commentUC := usecase.NewCommentUseCase(repo)
	authUC := usecase.NewAuthUseCase(repo, cfg)
	chatUC := usecase.NewChatUseCase(*repo, authUC)
	chatHandler := delivery.NewChatHandler(chatUC)
	// Инициализация обработчиков
	postHandler := delivery.NewPostHandler(*postUC, *commentUC)
	commentHandler := delivery.NewCommentHandler(*commentUC)
	authHandler := delivery.NewAuthHandler(*authUC)

	// Группа для аутентификации
	authGroup := router.Group("/auth")
	{
		authGroup.GET("/validate", authHandler.ValidateToken)
	}
	//chat group

	chatGroup := router.Group("/chat")
	{
		chatGroup.GET("/messages", chatHandler.GetMessages)
		chatGroup.GET("/ws", chatHandler.HandleWebSocket)
	}
	// Группа для постов
	postsGroup := router.Group("/posts")
	{
		postsGroup.GET("", postHandler.GetAllPosts)
		postsGroup.GET("/:id", postHandler.GetPostByID)

		protected := postsGroup.Group("")
		protected.Use(delivery.AuthMiddleware(cfg))
		{
			protected.POST("", postHandler.CreatePost)
			protected.DELETE("/:id", postHandler.DeletePost) // Переносим DELETE сюда
		}

		// Группа для комментариев
		commentsGroup := postsGroup.Group("/:id/comments")
		{
			commentsGroup.GET("", commentHandler.GetComments)

			protectedComments := commentsGroup.Group("")
			protectedComments.Use(delivery.AuthMiddleware(cfg))
			{
				protectedComments.POST("", commentHandler.CreateComment)
				protectedComments.DELETE("/:comment_id", commentHandler.DeleteComment)
			}
		}
	}

	log.Printf("Server is running on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

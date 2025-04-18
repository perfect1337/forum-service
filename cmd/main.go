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

	// Инициализация репозитория
	repo, err := repository.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	// Инициализация usecase
	postUC := usecase.NewPostUseCase(repo)

	// Инициализация HTTP сервера
	router := gin.Default()

	// Логирование запросов (должно быть первым middleware)
	router.Use(func(c *gin.Context) {
		log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Настройка CORS (должно быть перед обработчиками маршрутов)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Инициализация обработчиков
	postHandler := delivery.NewPostHandler(*postUC) // Исправлено здесь

	// Основной health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Группа постов
	posts := router.Group("/posts")
	{
		posts.GET("/", postHandler.GetAllPosts)
		posts.GET("/:id", postHandler.GetPostByID)

		// Защищенные роуты
		protected := posts.Group("")
		protected.Use(delivery.AuthMiddleware(cfg))
		{
			protected.POST("/", postHandler.CreatePost)
		}
	}
	// Запуск сервера (должен быть последним)
	log.Printf("Server is running on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

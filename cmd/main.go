package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/perfect1337/forum-service/docs"
	"github.com/perfect1337/forum-service/internal/config"
	grpcDelivery "github.com/perfect1337/forum-service/internal/delivery/grpcserver"
	delivery "github.com/perfect1337/forum-service/internal/delivery/http"
	"github.com/perfect1337/forum-service/internal/logger"
	forumPostProto "github.com/perfect1337/forum-service/internal/proto/post"
	"github.com/perfect1337/forum-service/internal/repository"
	"github.com/perfect1337/forum-service/internal/usecase"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Forum Service API
// @version 1.0
// @description API for managing forum posts, comments, and chat functionality
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"
func main() {
	cfg := config.Load()
	// Create context with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting auth service...")
	log.Infow("Loaded configuration",
		"server_port", cfg.Server.Port,
		"grpc_port", cfg.GRPC.Port,
		"log_level", cfg.Logger.LogLevel,
	)

	// Set Gin mode and disable console color
	if cfg.Logger.Development {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()

	// Initialize repository
	repo, err := repository.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	// Initialize use cases
	postUC := usecase.NewPostUseCase(repo, repo)
	commentUC := usecase.NewCommentUseCase(repo)
	authUC := usecase.NewAuthUseCase(*repo, cfg)
	chatUC := usecase.NewChatUseCase(repo, authUC)
	userUC := usecase.NewUserUseCase(repo)

	// Initialize gRPC connection to auth-service
	authAddr := os.Getenv("AUTH_SERVICE_GRPC_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50051"
	}

	authConn, err := grpc.DialContext(
		ctx,
		authAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	// Initialize gRPC server
	grpcSrv := grpc.NewServer()
	forumPostProto.RegisterPostServiceServer(
		grpcSrv,
		grpcDelivery.NewPostServer(postUC, authConn),
	)

	// Start gRPC server in goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.Postgres.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Initialize HTTP server
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(logger.GinLogger(log))
	router.Use(gin.Recovery())
	// Add Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize handlers
	postHandler := delivery.NewPostHandler(postUC, commentUC, userUC)
	commentHandler := delivery.NewCommentHandler(commentUC)
	authHandler := delivery.NewAuthHandler(authUC)
	chatHandler := delivery.NewChatHandler(chatUC)

	// Setup routes

	// Auth routes
	authGroup := router.Group("/auth")
	{
		authGroup.GET("/validate", authHandler.ValidateToken)
	}

	// Chat routes
	chat := router.Group("/chat")
	{
		chat.GET("/messages", chatHandler.GetMessages)
		chat.GET("/ws", chatHandler.HandleWebSocket)

		// Protected chat routes
		protected := chat.Group("")
		protected.Use(delivery.AuthMiddleware(cfg))
		{
			protected.POST("/messages", chatHandler.SendMessage)
		}
	}

	// Posts routes
	posts := router.Group("/posts")
	{
		posts.GET("", postHandler.GetAllPosts)
		posts.GET("/:id", postHandler.GetPostByID)

		// Protected routes
		protected := posts.Group("")
		protected.Use(delivery.AuthMiddleware(cfg))
		{
			protected.POST("", postHandler.CreatePost)
			protected.DELETE("/:id", postHandler.DeletePost)
		}

		// Comments routes
		comments := posts.Group("/:id/comments")
		{
			comments.GET("", commentHandler.GetComments)

			// Protected comments routes
			protectedComments := comments.Group("")
			protectedComments.Use(delivery.AuthMiddleware(cfg))
			{
				protectedComments.POST("", commentHandler.CreateComment)
				protectedComments.DELETE("/:comment_id", commentHandler.DeleteComment)
			}
		}
	}

	// Start HTTP server in goroutine
	go func() {
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully stop gRPC server
	grpcSrv.GracefulStop()

	// Cancel context
	cancel()

}

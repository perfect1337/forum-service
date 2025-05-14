package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	code := m.Run()

	os.Exit(code)
}

func TestHTTPServer(t *testing.T) {

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestGRPCServer(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	// Create a gRPC client connection
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

}

func TestAuthMiddleware(t *testing.T) {

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("username", "testuser")
		c.Next()
	})

	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, 1, userID)

		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/protected")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestCreatePost(t *testing.T) {
	router := gin.Default()
	router.POST("/posts", func(c *gin.Context) {
		var post struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, "Test Content", post.Content)

		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	server := httptest.NewServer(router)
	defer server.Close()
	postData := map[string]string{
		"title":   "Test Post",
		"content": "Test Content",
	}
	jsonData, _ := json.Marshal(postData)

	resp, err := http.Post(server.URL+"/posts", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "created", result["status"])
}

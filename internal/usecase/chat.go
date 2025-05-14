package usecase

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type AuthUseCase struct {
	repo      repository.Postgres
	secretKey []byte
}

func NewAuthUseCase(repo repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo:      repo,
		secretKey: []byte(cfg.Auth.SecretKey),
	}
}

// SecretKey returns the secret key used for signing and validating tokens.
func (uc *AuthUseCase) SecretKey() []byte {
	return uc.secretKey
}

// GenerateToken creates a JWT token with the specified claims.
func (uc *AuthUseCase) GenerateToken(userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.SecretKey())
}

type WebSocketConnection interface {
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error
	Close() error
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetPongHandler(h func(string) error)
	SetPingHandler(h func(string) error)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Subprotocol() string
	UnderlyingConn() net.Conn
}

// ParseToken parses and validates a JWT token, extracting the user ID and username from the claims.
func (uc *AuthUseCase) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.SecretKey(), nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, "", fmt.Errorf("invalid user_id in token")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return 0, "", fmt.Errorf("invalid username in token")
		}

		return int64(userID), username, nil
	}

	return 0, "", fmt.Errorf("invalid token")
}

type ChatRepository interface {
	SaveChatMessage(ctx context.Context, msg *entity.ChatMessage) error
	GetChatMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error)
	DeleteOldChatMessages(ctx context.Context, olderThan time.Duration) error
}

type ChatUseCase struct {
	repo   ChatRepository
	authUC AuthUseCaseInterface
	hub    *WebSocketHub
}

type AuthUseCaseInterface interface {
	SecretKey() []byte
	GenerateToken(userID int, username string) (string, error)
	ParseToken(tokenString string) (int64, string, error)
}

type WebSocketHub struct {
	clients         map[*WebSocketClient]bool
	broadcast       chan entity.ChatMessage
	register        chan *WebSocketClient
	unregister      chan *WebSocketClient
	maxConnections  int
	connectionCount int
	mutex           sync.Mutex
}

type ChatUseCaseInterface interface {
	SendMessage(ctx context.Context, message *entity.ChatMessage) error
	GetMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error)
	HandleWebSocket(conn WebSocketConnection) // Используем интерфейс вместо *websocket.Conn
}
type WebSocketClient struct {
	conn WebSocketConnection
	send chan entity.ChatMessage
}

func NewChatUseCase(repo ChatRepository, authUC AuthUseCaseInterface) *ChatUseCase {
	hub := newWebSocketHub(100)
	go hub.run()

	return &ChatUseCase{
		repo:   repo,
		authUC: authUC,
		hub:    hub,
	}
}
func (uc *ChatUseCase) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		err := uc.repo.DeleteOldChatMessages(ctx, 30*time.Minute)
		if err != nil {
			log.Printf("Error cleaning old messages: %v", err)
		}
	}
}
func newWebSocketHub(maxConnections int) *WebSocketHub {
	return &WebSocketHub{
		broadcast:      make(chan entity.ChatMessage),
		register:       make(chan *WebSocketClient),
		unregister:     make(chan *WebSocketClient),
		clients:        make(map[*WebSocketClient]bool),
		maxConnections: maxConnections,
	}
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (uc *ChatUseCase) HandleWebSocket(conn WebSocketConnection) {
	uc.hub.mutex.Lock()
	if uc.hub.connectionCount >= uc.hub.maxConnections {
		conn.WriteMessage(websocket.CloseMessage, []byte("too many connections"))
		conn.Close()
		uc.hub.mutex.Unlock()
		return
	}
	uc.hub.connectionCount++
	uc.hub.mutex.Unlock()

	client := &WebSocketClient{
		conn: conn,
		send: make(chan entity.ChatMessage, 256),
	}
	uc.hub.register <- client

	go func() {
		client.writePump()
		uc.hub.mutex.Lock()
		uc.hub.connectionCount--
		uc.hub.mutex.Unlock()
	}()
	client.readPump(uc)
}

func (c *WebSocketClient) readPump(uc *ChatUseCase) {
	defer func() {
		uc.hub.unregister <- c
		c.conn.Close()
	}()

	// При подключении очистим старые сообщения
	ctx := context.Background()
	_ = uc.repo.DeleteOldChatMessages(ctx, 30*time.Minute)
	for {
		var msg struct {
			Text  string `json:"text"`
			Token string `json:"token"`
		}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		userID, username, err := uc.authUC.ParseToken(msg.Token)
		if err != nil {
			log.Printf("Token validation error: %v", err)
			c.conn.WriteJSON(map[string]string{"error": "invalid token"})
			continue
		}

		if strings.TrimSpace(msg.Text) == "" {
			c.conn.WriteJSON(map[string]string{"error": "message cannot be empty"})
			continue
		}

		chatMsg := entity.ChatMessage{
			UserID:    int(userID), // Convert int64 to int for entity
			Author:    username,
			Text:      msg.Text,
			CreatedAt: time.Now(),
		}

		if err := uc.repo.SaveChatMessage(context.Background(), &chatMsg); err != nil {
			log.Printf("Error saving message: %v", err)
			c.conn.WriteJSON(map[string]string{"error": "failed to save message"})
			continue
		}

		uc.hub.broadcast <- chatMsg
	}
}

func (c *WebSocketClient) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.conn.WriteJSON(message)
	}
}

func (uc *ChatUseCase) SendMessage(ctx context.Context, message *entity.ChatMessage) error {
	if err := uc.repo.SaveChatMessage(ctx, message); err != nil {
		return err // Возвращаем ошибку из репозитория
	}
	uc.hub.broadcast <- *message
	return nil
}

func (uc *ChatUseCase) GetMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	// Сначала очистим старые сообщения перед получением
	err := uc.repo.DeleteOldChatMessages(ctx, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	return uc.repo.GetChatMessages(ctx, limit)
}

// auth_handler_test.go
package delivery

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test structures
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) SecretKey() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockAuthUseCase) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthUseCase) ParseToken(tokenString string) (int64, string, error) {
	args := m.Called(tokenString)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

// Test cases
func TestNewAuthHandler(t *testing.T) {
	mockAuthUC := new(MockAuthUseCase)
	handler := NewAuthHandler(mockAuthUC)
	assert.NotNil(t, handler)
}

func TestValidateToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthUC := new(MockAuthUseCase)
	mockAuthUC.On("SecretKey").Return([]byte("secret"))

	token, _ := jwt.New(jwt.SigningMethodHS256).SignedString([]byte("secret"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	handler := NewAuthHandler(mockAuthUC)
	handler.ValidateToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"valid":true`)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthUC := new(MockAuthUseCase)
	mockAuthUC.On("SecretKey").Return([]byte("secret"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token")

	handler := NewAuthHandler(mockAuthUC)
	handler.ValidateToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), `"valid":false`)
}

func TestValidateToken_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthUC := new(MockAuthUseCase)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := NewAuthHandler(mockAuthUC)
	handler.ValidateToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "token not provided")
}

func TestExtractToken_FromHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	assert.Equal(t, "token", extractToken(c))
}

func TestExtractToken_FromCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	assert.Equal(t, "cookie-token", extractToken(c))
}

func TestExtractToken_FromQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?token=query-token", nil)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	assert.Equal(t, "query-token", extractToken(c))
}

func TestExtractToken_NotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	assert.Equal(t, "", extractToken(c))
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем конфиг с тестовым секретным ключом
	cfg := &config.Config{
		Auth: struct {
			AccessTokenDuration  time.Duration
			RefreshTokenDuration time.Duration
			SecretKey            string
		}{
			SecretKey: "test-secret-key",
		},
	}

	// Создаем валидный JWT токен со всеми обязательными claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  1,                                // обязательный claim
		"username": "testuser",                       // обязательный claim
		"role":     "user",                           // опциональный claim
		"exp":      time.Now().Add(time.Hour).Unix(), // срок действия
	})
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	AuthMiddleware(cfg)(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, c.GetInt("user_id"))
	assert.Equal(t, "testuser", c.GetString("username"))
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Auth: struct {
			AccessTokenDuration  time.Duration
			RefreshTokenDuration time.Duration
			SecretKey            string
		}{
			SecretKey: "test-secret",
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token")

	AuthMiddleware(cfg)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("OPTIONS", "/", nil)

	AuthMiddleware(cfg)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractClaims_Valid(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id":  float64(1),
		"username": "test",
	}
	userID, username, _, err := extractClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)
	assert.Equal(t, "test", username)
}

func TestExtractClaims_InvalidUserID(t *testing.T) {
	claims := jwt.MapClaims{
		"username": "test",
	}
	_, _, _, err := extractClaims(claims)
	assert.Error(t, err)
}

func TestExtractClaims_InvalidUsername(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id": float64(1),
	}
	_, _, _, err := extractClaims(claims)
	assert.Error(t, err)
}

func TestAbortWithAuthError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	abortWithAuthError(c, "error", "code", "detail", "value")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"error"`)
	assert.Contains(t, w.Body.String(), `"detail":"value"`)
}

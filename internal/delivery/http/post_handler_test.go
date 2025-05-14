package delivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPostUseCase реализация мока для PostUseCase
// MockPostUseCase
type MockPostUseCase struct {
	mock.Mock
}

func (m *MockPostUseCase) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostUseCase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostUseCase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostUseCase) DeletePost(ctx context.Context, postID, userID int) error {
	args := m.Called(ctx, postID, userID)
	return args.Error(0)
}

// MockUserUseCase
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUseCase) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(map[int]*entity.User), args.Error(1)
}

// Проверка реализации интерфейсов
var _ usecase.PostUseCase = (*MockPostUseCase)(nil)

// ... остальные тестовые функции без изменений ...
func TestNewPostHandler(t *testing.T) {
	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	assert.NotNil(t, handler)
}

func TestPostHandler_CreatePost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	testUser := &entity.User{ID: 1, Username: "testuser"}
	mockUserUC.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)
	mockPostUC.On("CreatePost", mock.Anything, mock.AnythingOfType("*entity.Post")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)

	// Create a valid JSON body for the request
	postJSON := `{"title": "Test Post", "content": "This is a test post"}`
	c.Request = httptest.NewRequest("POST", "/posts", strings.NewReader(postJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.CreatePost(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockPostUC.AssertExpectations(t)
	mockUserUC.AssertExpectations(t)
}

func TestPostHandler_CreatePost_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/posts", nil)

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.CreatePost(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostHandler_CreatePost_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("POST", "/posts", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.CreatePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_GetPostByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	testPost := &entity.Post{ID: 1, UserID: 1}
	testUser := &entity.User{ID: 1, Username: "testuser"}
	testComments := []entity.Comment{{ID: 1, PostID: 1, UserID: 1}}

	mockPostUC.On("GetPostByID", mock.Anything, 1).Return(testPost, nil)
	mockUserUC.On("GetUserByID", mock.Anything, 1).Return(testUser, nil).Twice()
	mockCommentUC.On("GetCommentsByPostID", mock.Anything, 1).Return(testComments, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.GetPostByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"post":`)
	assert.Contains(t, w.Body.String(), `"comments":`)
	mockPostUC.AssertExpectations(t)
	mockCommentUC.AssertExpectations(t)
	mockUserUC.AssertExpectations(t)
}

func TestPostHandler_GetPostByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.GetPostByID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_GetPostByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Исправляем типы:
	// 1. Для контекста используем mock.Anything (он совместим с context.Context)
	// 2. Для ID используем int (как в реальном вызове)
	mockPostUC.On("GetPostByID", mock.Anything, 1).Return((*entity.Post)(nil), assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.GetPostByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockPostUC.AssertExpectations(t)
}
func TestPostHandler_GetAllPosts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	testPosts := []*entity.Post{{ID: 1, UserID: 1}, {ID: 2, UserID: 2}}
	testUser1 := &entity.User{ID: 1, Username: "user1"}
	testUser2 := &entity.User{ID: 2, Username: "user2"}

	mockPostUC.On("GetAllPosts", mock.Anything).Return(testPosts, nil)
	mockUserUC.On("GetUserByID", mock.Anything, 1).Return(testUser1, nil)
	mockUserUC.On("GetUserByID", mock.Anything, 2).Return(testUser2, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts", nil)

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.GetAllPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockPostUC.AssertExpectations(t)
	mockUserUC.AssertExpectations(t)
}

func TestPostHandler_GetAllPosts_WithComments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	testPosts := []*entity.Post{{ID: 1, UserID: 1}, {ID: 2, UserID: 2}}
	testUser1 := &entity.User{ID: 1, Username: "user1"}
	testUser2 := &entity.User{ID: 2, Username: "user2"}
	testComments1 := []entity.Comment{{ID: 1, PostID: 1, UserID: 1}}
	testComments2 := []entity.Comment{} // Пустой список комментариев для поста 2

	mockPostUC.On("GetAllPosts", mock.Anything).Return(testPosts, nil)
	mockUserUC.On("GetUserByID", mock.Anything, 1).Return(testUser1, nil).Twice()
	mockUserUC.On("GetUserByID", mock.Anything, 2).Return(testUser2, nil)
	// Настраиваем моки для обоих постов
	mockCommentUC.On("GetCommentsByPostID", mock.Anything, 1).Return(testComments1, nil)
	mockCommentUC.On("GetCommentsByPostID", mock.Anything, 2).Return(testComments2, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/posts?includeComments=true", nil)

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.GetAllPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockPostUC.AssertExpectations(t)
	mockUserUC.AssertExpectations(t)
	mockCommentUC.AssertExpectations(t)
}

func TestPostHandler_DeletePost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	mockPostUC.On("DeletePost", mock.Anything, 1, 1).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/posts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.DeletePost(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockPostUC.AssertExpectations(t)
}

func TestPostHandler_DeletePost_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/posts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Не устанавливаем user_id, чтобы симулировать неаутентифицированного пользователя

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.DeletePost(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPostHandler_DeletePost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/posts/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.DeletePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_DeletePost_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPostUC := new(MockPostUseCase)
	mockCommentUC := new(MockCommentUseCase)
	mockUserUC := new(MockUserUseCase)

	// Настраиваем моки
	mockPostUC.On("DeletePost", mock.Anything, 1, 1).Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/posts/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewPostHandler(mockPostUC, mockCommentUC, mockUserUC)
	handler.DeletePost(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockPostUC.AssertExpectations(t)
}

package delivery

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommentUseCase - мок для CommentUseCase
type MockCommentUseCase struct {
	mock.Mock
}

func (m *MockCommentUseCase) CreateComment(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentUseCase) GetCommentsByPostID(ctx context.Context, postID int) ([]entity.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]entity.Comment), args.Error(1)
}

func (m *MockCommentUseCase) DeleteComment(ctx context.Context, commentID, userID int) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

func TestCreateComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	// Create a test request with JSON body
	requestBody := `{"content": "Test comment"}`
	req, _ := http.NewRequest("POST", "/posts/1/comments", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Create expected comment
	expectedComment := &entity.Comment{
		PostID:  1,
		UserID:  1,
		Content: "Test comment",
	}

	// Set up mock
	mockCommentUC.On("CreateComment", mock.Anything, expectedComment).Return(nil)

	// Call the handler
	handler.CreateComment(c)

	// Check the response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify mock was called as expected
	mockCommentUC.AssertExpectations(t)
}

func TestGetComments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/1/comments", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	comments := []entity.Comment{
		{PostID: 1, UserID: 1, Content: "Comment 1"},
		{PostID: 1, UserID: 2, Content: "Comment 2"},
	}

	mockCommentUC.On("GetCommentsByPostID", mock.Anything, 1).Return(comments, nil)

	handler.GetComments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockCommentUC.AssertExpectations(t)
}

func TestDeleteComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/1", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	c.Params = gin.Params{gin.Param{Key: "comment_id", Value: "1"}}

	mockCommentUC.On("DeleteComment", mock.Anything, 1, 1).Return(nil)

	handler.DeleteComment(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockCommentUC.AssertExpectations(t)
}

func TestCreateCommentUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts/1/comments", strings.NewReader(`{"content": "Test"}`))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	handler.CreateComment(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateCommentInvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts/invalid/comments", strings.NewReader(`{"content": "Test"}`))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	handler.CreateComment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCommentsInvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/invalid/comments", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	handler.GetComments(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCommentInvalidCommentID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/invalid", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	c.Params = gin.Params{gin.Param{Key: "comment_id", Value: "invalid"}}

	handler.DeleteComment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCommentUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/1", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "comment_id", Value: "1"}}

	handler.DeleteComment(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteCommentNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCommentUC := new(MockCommentUseCase)
	handler := NewCommentHandler(mockCommentUC)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/1", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", 1)
	c.Params = gin.Params{gin.Param{Key: "comment_id", Value: "1"}}

	mockCommentUC.On("DeleteComment", mock.Anything, 1, 1).Return(sql.ErrNoRows)

	handler.DeleteComment(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockCommentUC.AssertExpectations(t)
}

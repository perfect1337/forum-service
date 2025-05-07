// internal/delivery/grpcserver/server_test.go
package grpcserver_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/perfect1337/forum-service/internal/delivery/grpcserver"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	userProto "github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
)

func TestPostServer_GetPostWithAuthor(t *testing.T) {
	mockPostUC := new(mocks.MockPostUseCase)
	mockUserClient := new(mocks.MockUserServiceClient)
	mockConn := &googlegrpc.ClientConn{}

	server := grpcserver.NewPostServer(
		mockPostUC,
		mockConn,
	)
	server.UserClient = mockUserClient

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		postID := 1
		authorID := 123
		expectedPost := &entity.Post{
			ID:      postID,
			Title:   "Test Post",
			Content: "Test Content",
			Author:  strconv.Itoa(authorID),
		}

		mockPostUC.On("GetPostByID", ctx, postID).Return(expectedPost, nil)
		mockUserClient.On("GetUsername", ctx, &userProto.UserRequest{UserId: int32(authorID)}, mock.Anything).
			Return(&userProto.UserResponse{Username: "testuser"}, nil)

		req := &postProto.PostRequest{PostId: int32(postID)}
		resp, err := server.GetPostWithAuthor(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resp.Title)
		assert.Equal(t, "testuser", resp.AuthorName)
		mockPostUC.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	// Остальные тест-кейсы остаются без изменений
	// ...
}

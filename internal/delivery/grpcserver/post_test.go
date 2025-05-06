// internal/delivery/grpcserver/server_test.go
package grpcserver_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/perfect1337/forum-service/internal/delivery/grpcserver"
	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	userProto "github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		mockUserClient.On("GetUsername", ctx, &userProto.UserRequest{UserId: int32(authorID)}).
			Return(&userProto.UserResponse{Username: "testuser"}, nil)

		req := &postProto.PostRequest{PostId: int32(postID)}
		resp, err := server.GetPostWithAuthor(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resp.Title)
		assert.Equal(t, "testuser", resp.AuthorName)
		mockPostUC.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	// Остальные тесты аналогично с исправлением обращения к полям ответа
	t.Run("PostNotFound", func(t *testing.T) {
		postID := 999
		mockPostUC.On("GetPostByID", ctx, postID).Return((*entity.Post)(nil), errors.New("not found"))

		req := &postProto.PostRequest{PostId: int32(postID)}
		_, err := server.GetPostWithAuthor(ctx, req)

		assert.Equal(t, codes.NotFound, status.Code(err))
		mockPostUC.AssertExpectations(t)
	})

	t.Run("InvalidAuthorFormat", func(t *testing.T) {
		postID := 2
		invalidPost := &entity.Post{
			ID:     postID,
			Author: "invalid",
		}

		mockPostUC.On("GetPostByID", ctx, postID).Return(invalidPost, nil)

		req := &postProto.PostRequest{PostId: int32(postID)}
		_, err := server.GetPostWithAuthor(ctx, req)

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mockPostUC.AssertExpectations(t)
	})

	t.Run("UserServiceError", func(t *testing.T) {
		postID := 3
		authorID := 456
		expectedPost := &entity.Post{
			ID:     postID,
			Author: strconv.Itoa(authorID),
		}

		mockPostUC.On("GetPostByID", ctx, postID).Return(expectedPost, nil)
		mockUserClient.On("GetUsername", ctx, &userProto.UserRequest{UserId: int32(authorID)}).
			Return((*userProto.UserResponse)(nil), errors.New("service unavailable"))

		req := &postProto.PostRequest{PostId: int32(postID)}
		_, err := server.GetPostWithAuthor(ctx, req)

		assert.Equal(t, codes.Internal, status.Code(err))
		mockPostUC.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})
}

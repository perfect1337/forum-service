package grpcserver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/perfect1337/forum-service/internal/delivery/grpcserver"
	"github.com/perfect1337/forum-service/internal/entity"
	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	userProto "github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockPostUsecase struct {
	mock.Mock
}

func (m *MockPostUsecase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Post), args.Error(1)
}

func (m *MockPostUsecase) CreatePost(ctx context.Context, post *entity.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostUsecase) GetAllPosts(ctx context.Context) ([]*entity.Post, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Post), args.Error(1)
}

func (m *MockPostUsecase) DeletePost(ctx context.Context, postID, userID int) error {
	args := m.Called(ctx, postID, userID)
	return args.Error(0)
}

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUsername(ctx context.Context, in *userProto.UserRequest, opts ...grpc.CallOption) (*userProto.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userProto.UserResponse), args.Error(1)
}

func TestPostServer_GetPostWithAuthor(t *testing.T) {
	tests := []struct {
		name           string
		req            *postProto.PostRequest
		mockPostSetup  func(*MockPostUsecase)
		mockUserSetup  func(*MockUserClient)
		expectedResp   *postProto.PostResponse
		expectedErr    error
		expectedErrMsg string
	}{
		{
			name: "Success",
			req:  &postProto.PostRequest{PostId: 1},
			mockPostSetup: func(m *MockPostUsecase) {
				m.On("GetPostByID", mock.Anything, 1).
					Return(&entity.Post{
						ID:      1,
						Title:   "Test Post",
						Content: "Test Content",
						Author:  "123",
					}, nil)
			},
			mockUserSetup: func(m *MockUserClient) {
				m.On("GetUsername", mock.Anything, &userProto.UserRequest{UserId: 123}).
					Return(&userProto.UserResponse{Username: "testuser"}, nil)
			},
			expectedResp: &postProto.PostResponse{
				Id:         1,
				Title:      "Test Post",
				Content:    "Test Content",
				AuthorName: "testuser",
			},
		},
		{
			name: "PostNotFound",
			req:  &postProto.PostRequest{PostId: 2},
			mockPostSetup: func(m *MockPostUsecase) {
				m.On("GetPostByID", mock.Anything, 2).
					Return(nil, errors.New("post not found"))
			},
			mockUserSetup:  func(m *MockUserClient) {},
			expectedErr:    status.Error(codes.NotFound, "post not found: post not found"),
			expectedErrMsg: "post not found: post not found",
		},
		{
			name: "InvalidAuthorIDFormat",
			req:  &postProto.PostRequest{PostId: 3},
			mockPostSetup: func(m *MockPostUsecase) {
				m.On("GetPostByID", mock.Anything, 3).
					Return(&entity.Post{
						ID:      3,
						Title:   "Invalid Author",
						Content: "Test Content",
						Author:  "invalid",
					}, nil)
			},
			mockUserSetup:  func(m *MockUserClient) {},
			expectedErr:    status.Error(codes.InvalidArgument, "invalid author ID format: strconv.Atoi: parsing \"invalid\": invalid syntax"),
			expectedErrMsg: "invalid author ID format",
		},
		{
			name: "UserServiceError",
			req:  &postProto.PostRequest{PostId: 4},
			mockPostSetup: func(m *MockPostUsecase) {
				m.On("GetPostByID", mock.Anything, 4).
					Return(&entity.Post{
						ID:      4,
						Title:   "Test Post",
						Content: "Test Content",
						Author:  "456",
					}, nil)
			},
			mockUserSetup: func(m *MockUserClient) {
				m.On("GetUsername", mock.Anything, &userProto.UserRequest{UserId: 456}).
					Return(nil, errors.New("user service unavailable"))
			},
			expectedErr:    status.Error(codes.Internal, "failed to get username: user service unavailable"),
			expectedErrMsg: "failed to get username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postUsecase := new(MockPostUsecase)
			userClient := new(MockUserClient)

			tt.mockPostSetup(postUsecase)
			tt.mockUserSetup(userClient)

			server := grpcserver.NewPostServer(postUsecase, nil)
			server.UserClient = userClient

			resp, err := server.GetPostWithAuthor(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				statusErr, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErr.(interface{ GRPCStatus() *status.Status }).GRPCStatus().Code(), statusErr.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}

			postUsecase.AssertExpectations(t)
			userClient.AssertExpectations(t)
		})
	}
}

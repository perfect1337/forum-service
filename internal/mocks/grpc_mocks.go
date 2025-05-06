package mocks

import (
	"context"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockPostUseCase struct {
	mock.Mock
}
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUsername(ctx context.Context, in *user.UserRequest, opts ...grpc.CallOption) (*user.UserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*user.UserResponse), args.Error(1)
}
func (m *MockPostUseCase) GetPostByID(ctx context.Context, id int) (*entity.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Post), args.Error(1)
}

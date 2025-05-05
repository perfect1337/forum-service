package grpc

import (
	"context"

	"github.com/perfect1337/forum-service/internal/proto/post"
	"github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/perfect1337/forum-service/internal/usecase"
	"google.golang.org/grpc"
)

type PostServer struct {
	post.UnimplementedPostServiceServer
	postUsecase usecase.Post
	userClient  user.UserServiceClient
}

func NewPostServer(p usecase.Post, conn *grpc.ClientConn) *PostServer {
	return &PostServer{
		postUsecase: p,
		userClient:  user.NewUserServiceClient(conn),
	}
}

func (s *PostServer) GetPostAuthor(ctx context.Context, req *post.PostRequest) (*user.UserResponse, error) {
	// 1. Получаем пост из репозитория
	p, err := s.postUsecase.GetByID(ctx, int(req.PostId))
	if err != nil {
		return nil, err
	}

	// 2. Запрашиваем имя пользователя через gRPC
	return s.userClient.GetUsername(ctx, &user.UserRequest{UserId: int32(p.AuthorID)})
}

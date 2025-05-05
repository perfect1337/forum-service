package grpc

import (
	"context"

	authProto "github.com/perfect1337/auth-service/internal/proto"
	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	"github.com/perfect1337/forum-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostServer struct {
	postProto.UnimplementedPostServiceServer
	postUsecase usecase.PostUseCase
	authClient  authProto.UserServiceClient
}

func NewPostServer(postUC usecase.PostUseCase, authConn *grpc.ClientConn) *PostServer {
	return &PostServer{
		postUsecase: postUC,
		authClient:  authProto.NewUserServiceClient(authConn),
	}
}

func (s *PostServer) GetPostWithAuthor(ctx context.Context, req *postProto.PostRequest) (*postProto.PostResponse, error) {
	// 1. Получаем пост из репозитория
	post, err := s.postUsecase.GetByID(ctx, int(req.GetPostId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	// 2. Получаем имя пользователя через GetUsername
	usernameResp, err := s.authClient.GetUsername(ctx, &authProto.UserRequest{
		UserId: int32(post.AuthorID),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get username: %v", err)
	}

	// 3. Формируем ответ
	return &postProto.PostResponse{
		Id:         int32(post.ID),
		Title:      post.Title,
		Content:    post.Content,
		AuthorName: usernameResp.GetUsername(), // Используем полученное имя
	}, nil
}

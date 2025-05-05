package grpc

import (
	"context"

	userProto "github.com/perfect1337/auth-service/internal/proto"
	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	"github.com/perfect1337/forum-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostServer struct {
	postProto.UnimplementedPostServiceServer
	postUsecase usecase.PostUseCase
	userClient  userProto.UserServiceClient
}

func NewPostServer(postUC usecase.PostUseCase, userConn *grpc.ClientConn) *PostServer {
	return &PostServer{
		postUsecase: postUC,
		userClient:  userProto.NewUserServiceClient(userConn),
	}
}

func (s *PostServer) GetPostWithAuthor(ctx context.Context, req *postProto.PostRequest) (*postProto.PostResponse, error) {
	// 1. Получаем пост из репозитория
	post, err := s.postUsecase.GetByID(ctx, int(req.GetPostId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	// 2. Делаем gRPC вызов к auth-service для получения имени автора
	usernameResp, err := s.userClient.GetUsername(ctx, &userProto.UserRequest{
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
		AuthorName: usernameResp.GetUsername(),
	}, nil
}

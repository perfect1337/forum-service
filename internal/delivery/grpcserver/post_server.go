// internal/delivery/grpcserver/server.go
package grpcserver

import (
	"context"
	"strconv"

	postProto "github.com/perfect1337/forum-service/internal/proto/post"
	userProto "github.com/perfect1337/forum-service/internal/proto/user"
	"github.com/perfect1337/forum-service/internal/usecase"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostServer struct {
	postProto.UnimplementedPostServiceServer
	postUsecase usecase.PostUseCase
	UserClient  userProto.UserServiceClient // Публичное поле
}

func NewPostServer(postUC usecase.PostUseCase, userConn *googlegrpc.ClientConn) *PostServer {
	return &PostServer{
		postUsecase: postUC,
		UserClient:  userProto.NewUserServiceClient(userConn), // Исправлено имя поля
	}
}

func (s *PostServer) GetPostWithAuthor(ctx context.Context, req *postProto.PostRequest) (*postProto.PostResponse, error) {
	post, err := s.postUsecase.GetPostByID(ctx, int(req.GetPostId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	authorID, err := strconv.Atoi(post.Author)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid author ID format: %v", err)
	}

	usernameResp, err := s.UserClient.GetUsername(ctx, &userProto.UserRequest{
		UserId: int32(authorID),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get username: %v", err)
	}

	return &postProto.PostResponse{
		Id:         int32(post.ID),
		Title:      post.Title,
		Content:    post.Content,
		AuthorName: usernameResp.GetUsername(),
	}, nil
}

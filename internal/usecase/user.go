package usecase

import (
	"context"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return uc.userRepo.GetUserByID(ctx, id)
}

func (uc *UserUseCase) GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	return uc.userRepo.GetUsersByIDs(ctx, ids)
}

type UserUseCaseInterface interface {
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetUsersByIDs(ctx context.Context, ids []int) (map[int]*entity.User, error)
}

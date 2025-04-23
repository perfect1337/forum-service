package usecase

import (
	"context"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/repository"
)

type ChatUseCase struct {
	repo *repository.Postgres
}

func NewChatUseCase(repo *repository.Postgres) *ChatUseCase {
	return &ChatUseCase{repo: repo}
}

func (uc *ChatUseCase) SendMessage(ctx context.Context, message *entity.ChatMessage) error {
	return uc.repo.CreateChatMessage(ctx, message)
}

func (uc *ChatUseCase) GetMessages(ctx context.Context, limit int) ([]entity.ChatMessage, error) {
	return uc.repo.GetChatMessages(ctx, limit)
}

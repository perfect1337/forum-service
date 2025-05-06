package usecase_test

import (
	"context"
	"testing"

	"github.com/perfect1337/forum-service/internal/entity"
	"github.com/perfect1337/forum-service/internal/mocks"
	"github.com/perfect1337/forum-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommentUseCase_CreateComment(t *testing.T) {
	repo := new(mocks.MockCommentRepository)
	uc := usecase.NewCommentUseCase(repo)

	t.Run("Success", func(t *testing.T) {
		comment := &entity.Comment{Content: "test"}
		repo.On("CreateComment", mock.Anything, comment).Return(nil)

		err := uc.CreateComment(context.Background(), comment)
		assert.NoError(t, err)
	})
}

package review

import (
	"context"
	"testing"

	"guidely-app/internal/review/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 1,
		Rating:  5,
		Comment: "Great!",
	}

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, r *models.Review) error {
		r.ID = 1
		return nil
	})

	review, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), review.ID)
	assert.Equal(t, int16(5), review.Rating)
}

func TestService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, nil)

	err := svc.Delete(context.Background(), 1, 1)
	assert.EqualError(t, err, "review not found")
}

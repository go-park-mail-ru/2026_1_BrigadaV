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
	input := CreateReviewInput{UserID: 1, PlaceID: 1, Rating: 5, Comment: "Great"}
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, r *models.Review) error {
		r.ID = 1
		return nil
	})
	review, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), review.ID)
}

func TestService_Create_InvalidRating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)
	input := CreateReviewInput{UserID: 1, PlaceID: 1, Rating: 6, Comment: "Great"}
	_, err := svc.Create(context.Background(), input)
	assert.EqualError(t, err, "rating must be between 1 and 5")
}

func TestService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)
	review := &models.Review{ID: 1, UserID: 1}
	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(review, nil)
	repo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)
	err := svc.Delete(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestService_Delete_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)
	review := &models.Review{ID: 1, UserID: 2}
	repo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(review, nil)
	err := svc.Delete(context.Background(), 1, 1)
	assert.EqualError(t, err, "not authorized")
}

func TestService_GetByPlaceIDWithAuthor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockReviewRepository(ctrl)
	svc := NewService(repo)

	var review models.ReviewWithAuthor
	review.ID = 1
	review.Rating = 5
	review.Comment = "Great"
	review.Author.ID = 1
	review.Author.Nickname = "john"

	repo.EXPECT().GetByPlaceIDWithAuthor(gomock.Any(), uint64(1)).Return([]models.ReviewWithAuthor{review}, nil)
	result, err := svc.GetByPlaceIDWithAuthor(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

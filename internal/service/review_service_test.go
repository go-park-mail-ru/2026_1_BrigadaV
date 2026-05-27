package service

import (
	"context"
	"errors"
	"testing"

	"guidely-app/internal/repository/mocks"
	"guidely-app/pkg/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReviewService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	input := CreateReviewInput{
		UserID:    1,
		PlaceID:   10,
		Rating:    5,
		Comment:   "Great!",
		VisitDate: nil,
	}
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, r *models.Review) error {
		r.ID = 100
		return nil
	})

	review, err := svc.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, uint64(100), review.ID)
	assert.Equal(t, int16(5), review.Rating)
}

func TestReviewService_Create_InvalidRating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 10,
		Rating:  6,
	}
	review, err := svc.Create(context.Background(), input)
	assert.Error(t, err)
	assert.Nil(t, review)
	assert.Equal(t, "rating must be between 1 and 5", err.Error())
}

func TestReviewService_Create_RatingZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 10,
		Rating:  0,
	}
	review, err := svc.Create(context.Background(), input)
	assert.Error(t, err)
	assert.Nil(t, review)
}

func TestReviewService_Create_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 10,
		Rating:  3,
	}
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("duplicate key"))

	review, err := svc.Create(context.Background(), input)
	assert.Error(t, err)
	assert.Nil(t, review)
	assert.Equal(t, "duplicate key", err.Error())
}

func TestReviewService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	review := &models.Review{ID: 5, UserID: 10}
	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(5)).Return(review, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), uint64(5)).Return(nil)

	err := svc.Delete(context.Background(), 10, 5)
	assert.NoError(t, err)
}

func TestReviewService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(5)).Return(nil, nil)

	err := svc.Delete(context.Background(), 10, 5)
	assert.Error(t, err)
	assert.Equal(t, "review not found", err.Error())
}

func TestReviewService_Delete_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	review := &models.Review{ID: 5, UserID: 20}
	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(5)).Return(review, nil)

	err := svc.Delete(context.Background(), 10, 5)
	assert.Error(t, err)
	assert.Equal(t, "not authorized", err.Error())
}

func TestReviewService_Delete_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(5)).Return(nil, errors.New("db error"))

	err := svc.Delete(context.Background(), 10, 5)
	assert.Error(t, err)
}

func TestReviewService_Delete_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewReviewService(mockRepo)

	review := &models.Review{ID: 5, UserID: 10}
	mockRepo.EXPECT().GetByID(gomock.Any(), uint64(5)).Return(review, nil)
	mockRepo.EXPECT().Delete(gomock.Any(), uint64(5)).Return(errors.New("db delete error"))

	err := svc.Delete(context.Background(), 10, 5)
	assert.Error(t, err)
}

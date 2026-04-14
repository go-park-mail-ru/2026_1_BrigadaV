package service

import (
	"context"
	"testing"

	"guidely-app/internal/models"
	"guidely-app/internal/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReviewService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	service := NewReviewService(mockReviewRepo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 1,
		Rating:  5,
		Comment: "Great!",
	}

	mockReviewRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	review, err := service.Create(context.Background(), input)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, int16(5), review.Rating)
}

func TestReviewService_Create_InvalidRating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	service := NewReviewService(mockReviewRepo)

	input := CreateReviewInput{
		UserID:  1,
		PlaceID: 1,
		Rating:  6,
		Comment: "Great!",
	}

	review, err := service.Create(context.Background(), input)
	assert.Error(t, err)
	assert.Nil(t, review)
	assert.Equal(t, "rating must be between 1 and 5", err.Error())
}

func TestReviewService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	service := NewReviewService(mockReviewRepo)

	review := &models.Review{ID: 1, UserID: 1}
	mockReviewRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(review, nil)
	mockReviewRepo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)

	err := service.Delete(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestReviewService_Delete_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	service := NewReviewService(mockReviewRepo)

	review := &models.Review{ID: 1, UserID: 2}
	mockReviewRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(review, nil)

	err := service.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "not authorized", err.Error())
}

func TestReviewService_Delete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	service := NewReviewService(mockReviewRepo)

	mockReviewRepo.EXPECT().GetByID(gomock.Any(), uint64(1)).Return(nil, nil)

	err := service.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Equal(t, "review not found", err.Error())
}

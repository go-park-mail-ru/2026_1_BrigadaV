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

func TestPlaceService_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	expectedPlaces := []models.Place{{ID: 1, Name: "Eiffel Tower"}}
	mockPlaceRepo.EXPECT().GetAll(gomock.Any()).Return(expectedPlaces, nil)

	places, err := svc.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, places, 1)
	assert.Equal(t, "Eiffel Tower", places[0].Name)
}

func TestPlaceService_GetAll_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	mockPlaceRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	places, err := svc.GetAll(context.Background())
	assert.Error(t, err)
	assert.Nil(t, places)
}

func TestPlaceService_GetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	expected := &models.PlaceWithRating{ID: 1, Name: "Eiffel Tower", Rating: 4.5}
	mockPlaceRepo.EXPECT().GetWithRatingAndLike(gomock.Any(), uint64(1), uint64(1)).Return(expected, nil)

	result, err := svc.GetDetails(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Eiffel Tower", result.Name)
}

func TestPlaceService_GetByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	places := []models.Place{{ID: 1, Name: "Hotel"}, {ID: 2, Name: "Motel"}}
	mockPlaceRepo.EXPECT().GetByCategory(gomock.Any(), uint64(1)).Return(places, nil)
	result, err := svc.GetByCategory(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestPlaceService_GetByCategory_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	mockPlaceRepo.EXPECT().GetByCategory(gomock.Any(), uint64(1)).Return(nil, errors.New("db error"))
	_, err := svc.GetByCategory(context.Background(), 1)
	assert.Error(t, err)
}

func TestPlaceService_Search(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	places := []models.Place{{ID: 1, Name: "Eiffel Tower"}}
	mockPlaceRepo.EXPECT().Search(gomock.Any(), "eiffel").Return(places, nil)
	result, err := svc.Search(context.Background(), "eiffel")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestPlaceService_Search_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	mockPlaceRepo.EXPECT().Search(gomock.Any(), "eiffel").Return(nil, errors.New("db error"))
	_, err := svc.Search(context.Background(), "eiffel")
	assert.Error(t, err)
}

func TestPlaceService_IsPlaceInTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	mockReviewRepo := mocks.NewMockReviewRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo, mockReviewRepo)

	mockPlaceRepo.EXPECT().IsPlaceInTrip(gomock.Any(), uint64(1), uint64(2)).Return(true, nil)
	inTrip, err := svc.IsPlaceInTrip(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.True(t, inTrip)
}

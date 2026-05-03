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
	svc := NewPlaceService(mockPlaceRepo)

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
	svc := NewPlaceService(mockPlaceRepo)

	mockPlaceRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	places, err := svc.GetAll(context.Background())
	assert.Error(t, err)
	assert.Nil(t, places)
}

func TestPlaceService_GetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPlaceRepo := mocks.NewMockPlaceRepository(ctrl)
	svc := NewPlaceService(mockPlaceRepo)

	expected := &models.PlaceWithRating{ID: 1, Name: "Eiffel Tower", Rating: 4.5}
	mockPlaceRepo.EXPECT().GetWithRatingAndLike(gomock.Any(), uint64(1), uint64(1)).Return(expected, nil)

	result, err := svc.GetDetails(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Eiffel Tower", result.Name)
}

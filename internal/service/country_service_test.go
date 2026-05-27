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

func TestCountryService_GetAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	expected := []models.Country{
		{ID: 1, Name: "France"},
		{ID: 2, Name: "Italy"},
	}
	mockRepo.EXPECT().GetAll(gomock.Any()).Return(expected, nil)

	countries, err := svc.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Equal(t, "France", countries[0].Name)
}

func TestCountryService_GetAll_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	countries, err := svc.GetAll(context.Background())
	assert.Error(t, err)
	assert.Nil(t, countries)
}

func TestCountryService_GetWithLocalities_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	countries := []models.Country{
		{ID: 1, Name: "France"},
		{ID: 2, Name: "Italy"},
	}
	localities := []models.Locality{
		{ID: 10, Name: "Paris", Country: "France"},
		{ID: 11, Name: "Lyon", Country: "France"},
	}

	mockRepo.EXPECT().GetAll(gomock.Any()).Return(countries, nil)
	mockRepo.EXPECT().GetLocalitiesByCountryID(gomock.Any(), uint64(1)).Return(localities, nil)

	country, locs, err := svc.GetWithLocalities(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, country)
	assert.Equal(t, "France", country.Name)
	assert.Len(t, locs, 2)
}

func TestCountryService_GetWithLocalities_CountryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	countries := []models.Country{{ID: 1, Name: "France"}}
	mockRepo.EXPECT().GetAll(gomock.Any()).Return(countries, nil)

	country, locs, err := svc.GetWithLocalities(context.Background(), 999)
	assert.NoError(t, err)
	assert.Nil(t, country)
	assert.Nil(t, locs)
}

func TestCountryService_GetWithLocalities_GetLocalitiesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	countries := []models.Country{{ID: 1, Name: "France"}}
	mockRepo.EXPECT().GetAll(gomock.Any()).Return(countries, nil)
	mockRepo.EXPECT().GetLocalitiesByCountryID(gomock.Any(), uint64(1)).Return(nil, errors.New("db error"))

	country, locs, err := svc.GetWithLocalities(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, country)
	assert.Nil(t, locs)
}

func TestCountryService_GetWithLocalities_GetAllError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCountryRepository(ctrl)
	svc := NewCountryService(mockRepo)

	mockRepo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	country, locs, err := svc.GetWithLocalities(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, country)
	assert.Nil(t, locs)
}

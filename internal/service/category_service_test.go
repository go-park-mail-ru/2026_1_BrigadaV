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

func TestCategoryService_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockCategoryRepository(ctrl)
	svc := NewCategoryService(repo)

	expected := []models.Category{
		{ID: 1, Name: "Отель", ApplicableTypes: []string{"hotel"}},
	}
	repo.EXPECT().GetAll(gomock.Any()).Return(expected, nil)

	categories, err := svc.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, categories, 1)
	assert.Equal(t, "Отель", categories[0].Name)
}

func TestCategoryService_GetAll_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockCategoryRepository(ctrl)
	svc := NewCategoryService(repo)

	repo.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("db error"))

	_, err := svc.GetAll(context.Background())
	assert.Error(t, err)
}

func TestCategoryService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockCategoryRepository(ctrl)
	svc := NewCategoryService(repo)

	category := &models.Category{Name: "Test", ApplicableTypes: []string{"attraction"}}
	repo.EXPECT().Create(gomock.Any(), category).Return(nil)

	err := svc.Create(context.Background(), category)
	assert.NoError(t, err)
}

func TestCategoryService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockCategoryRepository(ctrl)
	svc := NewCategoryService(repo)

	repo.EXPECT().Delete(gomock.Any(), uint64(1)).Return(nil)

	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

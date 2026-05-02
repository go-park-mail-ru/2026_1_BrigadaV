package service

import (
	"context"
	"guidely-app/internal/repository"
	"guidely-app/pkg/models"
)

type categoryServiceImpl struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryServiceImpl{repo: repo}
}

func (s *categoryServiceImpl) GetAll(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *categoryServiceImpl) GetByID(ctx context.Context, id uint64) (*models.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryServiceImpl) Create(ctx context.Context, c *models.Category) error {
	return s.repo.Create(ctx, c)
}

func (s *categoryServiceImpl) Update(ctx context.Context, c *models.Category) error {
	return s.repo.Update(ctx, c)
}

func (s *categoryServiceImpl) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAll(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *CategoryService) GetByID(ctx context.Context, id uint64) (*models.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) Create(ctx context.Context, c *models.Category) error {
	return s.repo.Create(ctx, c)
}

func (s *CategoryService) Update(ctx context.Context, c *models.Category) error {
	return s.repo.Update(ctx, c)
}

func (s *CategoryService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

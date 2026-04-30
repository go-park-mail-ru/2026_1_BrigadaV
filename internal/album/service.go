package album

import (
	"context"
	"guidely-app/internal/album/repository"
	"guidely-app/pkg/models"
)

type Service struct {
	repo repository.AlbumRepository
}

func NewService(repo repository.AlbumRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, a *models.Album) error {
	return s.repo.Create(ctx, a)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return s.repo.GetByTrip(ctx, tripID)
}

func (s *Service) Update(ctx context.Context, a *models.Album) error {
	return s.repo.Update(ctx, a)
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

// albumServiceImpl – приватная реализация интерфейса AlbumService
type albumServiceImpl struct {
	repo repository.AlbumRepository
}

func NewAlbumService(repo repository.AlbumRepository) AlbumService {
	return &albumServiceImpl{repo: repo}
}

func (s *albumServiceImpl) Create(ctx context.Context, album *models.Album) error {
	return s.repo.Create(ctx, album)
}

func (s *albumServiceImpl) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *albumServiceImpl) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return s.repo.GetByTrip(ctx, tripID)
}

func (s *albumServiceImpl) Update(ctx context.Context, album *models.Album) error {
	return s.repo.Update(ctx, album)
}

func (s *albumServiceImpl) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

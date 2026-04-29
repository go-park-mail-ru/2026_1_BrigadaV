package service

import (
	"context"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
)

type AlbumService struct {
	repo *repository.AlbumRepo
}

func NewAlbumService(repo *repository.AlbumRepo) *AlbumService {
	return &AlbumService{repo: repo}
}

func (s *AlbumService) Create(ctx context.Context, album *models.Album) error {
	return s.repo.Create(ctx, album)
}

func (s *AlbumService) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AlbumService) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return s.repo.GetByTrip(ctx, tripID)
}

func (s *AlbumService) Update(ctx context.Context, album *models.Album) error {
	return s.repo.Update(ctx, album)
}

func (s *AlbumService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

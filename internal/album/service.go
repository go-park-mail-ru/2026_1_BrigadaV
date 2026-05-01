package album

import (
	"context"
	"guidely-app/internal/album/repository"
	"guidely-app/pkg/models"
)

type AlbumService interface {
	Create(ctx context.Context, a *models.Album) error
	GetByID(ctx context.Context, id uint64) (*models.Album, error)
	GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error)
	Update(ctx context.Context, a *models.Album) error
	Delete(ctx context.Context, id uint64) error
}

type albumService struct {
	repo repository.AlbumRepository
}

func NewService(repo repository.AlbumRepository) AlbumService {
	return &albumService{repo: repo}
}

func (s *albumService) Create(ctx context.Context, a *models.Album) error {
	return s.repo.Create(ctx, a)
}
func (s *albumService) GetByID(ctx context.Context, id uint64) (*models.Album, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *albumService) GetByTrip(ctx context.Context, tripID uint64) ([]models.Album, error) {
	return s.repo.GetByTrip(ctx, tripID)
}
func (s *albumService) Update(ctx context.Context, a *models.Album) error {
	return s.repo.Update(ctx, a)
}
func (s *albumService) Delete(ctx context.Context, id uint64) error { return s.repo.Delete(ctx, id) }
